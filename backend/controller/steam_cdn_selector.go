package controller

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"l4d2-manager-next/logic"
)

const (
	steamCDNIPListAPI       = "https://lytvpk-get-ips.laoyutang.cn/?domain=cdn.steamusercontent.com&format=entries"
	steamCDNSelectionTTL    = 30 * time.Minute
	steamCDNFailureCooldown = 10 * time.Minute
	steamCDNRetryDelay      = time.Minute
	steamCDNCandidateLimit  = 32
	steamCDNSpeedTestLimit  = 5
	steamCDNSpeedTestBytes  = 512 * 1024
)

var defaultSteamCDNIPs = []string{
	"23.59.72.59",
	"23.59.72.42",
	"23.206.175.162",
	"23.206.175.170",
	"23.55.51.221",
	"23.67.75.74",
}

type steamCDNIPOption struct {
	IP       string `json:"ip"`
	Category string `json:"category"`
}

type steamCDNSelection struct {
	IP       string
	Category string
	Speed    float64
}

type steamCDNSelectorSnapshot struct {
	Mode           string    `json:"steam_cdn_mode"`
	ActiveIP       string    `json:"steam_cdn_active_ip"`
	Category       string    `json:"steam_cdn_category,omitempty"`
	SpeedMBps      float64   `json:"steam_cdn_speed_mbps,omitempty"`
	SelectedAt     time.Time `json:"steam_cdn_selected_at,omitempty"`
	Selecting      bool      `json:"steam_cdn_selecting"`
	CandidateCount int       `json:"steam_cdn_candidate_count"`
}

type steamCDNSelector struct {
	mu             sync.Mutex
	selected       steamCDNSelection
	selectedAt     time.Time
	lastAttemptAt  time.Time
	selecting      bool
	wait           chan struct{}
	failedUntil    map[string]time.Time
	candidateCount int
	now            func() time.Time
	choose         func(context.Context, string, map[string]time.Time) (steamCDNSelection, int)
}

var globalSteamCDNSelector = newSteamCDNSelector()

func newSteamCDNSelector() *steamCDNSelector {
	return &steamCDNSelector{
		failedUntil: make(map[string]time.Time),
		now:         time.Now,
		choose:      chooseBestSteamCDNIP,
	}
}

func (s *steamCDNSelector) Select(ctx context.Context, testURL string) string {
	if fixedIP := logic.GetSteamCDNIP(); fixedIP != "" {
		return fixedIP
	}
	parsed, err := url.Parse(testURL)
	if err != nil || !isSteamCDNDownloadHost(parsed.Hostname()) {
		return ""
	}

	for {
		s.mu.Lock()
		now := s.now()
		if s.selected.IP != "" && now.Sub(s.selectedAt) < steamCDNSelectionTTL && !now.Before(s.failedUntil[s.selected.IP]) {
			ip := s.selected.IP
			s.mu.Unlock()
			return ip
		}
		if s.selecting {
			wait := s.wait
			s.mu.Unlock()
			select {
			case <-ctx.Done():
				return ""
			case <-wait:
				continue
			}
		}
		if s.selected.IP == "" && !s.lastAttemptAt.IsZero() && now.Sub(s.lastAttemptAt) < steamCDNRetryDelay {
			s.mu.Unlock()
			return ""
		}

		s.selecting = true
		s.wait = make(chan struct{})
		blocked := make(map[string]time.Time, len(s.failedUntil))
		for ip, until := range s.failedUntil {
			if now.Before(until) {
				blocked[ip] = until
			} else {
				delete(s.failedUntil, ip)
			}
		}
		s.mu.Unlock()

		selection, candidateCount := s.choose(ctx, testURL, blocked)

		s.mu.Lock()
		s.lastAttemptAt = s.now()
		s.candidateCount = candidateCount
		if selection.IP != "" {
			s.selected = selection
			s.selectedAt = s.lastAttemptAt
		} else {
			s.selected = steamCDNSelection{}
			s.selectedAt = time.Time{}
		}
		s.selecting = false
		close(s.wait)
		s.mu.Unlock()
		return selection.IP
	}
}

func (s *steamCDNSelector) MarkFailed(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == logic.GetSteamCDNIP() {
		return
	}
	s.mu.Lock()
	s.failedUntil[ip] = s.now().Add(steamCDNFailureCooldown)
	if s.selected.IP == ip {
		s.selected = steamCDNSelection{}
		s.selectedAt = time.Time{}
		s.lastAttemptAt = time.Time{}
	}
	s.mu.Unlock()
}

func (s *steamCDNSelector) Reset() {
	s.mu.Lock()
	s.selected = steamCDNSelection{}
	s.selectedAt = time.Time{}
	s.lastAttemptAt = time.Time{}
	s.failedUntil = make(map[string]time.Time)
	s.candidateCount = 0
	s.mu.Unlock()
}

func (s *steamCDNSelector) Snapshot() steamCDNSelectorSnapshot {
	fixedIP := logic.GetSteamCDNIP()
	s.mu.Lock()
	defer s.mu.Unlock()
	if fixedIP != "" {
		return steamCDNSelectorSnapshot{Mode: "fixed", ActiveIP: fixedIP, CandidateCount: s.candidateCount}
	}
	return steamCDNSelectorSnapshot{
		Mode:           "auto",
		ActiveIP:       s.selected.IP,
		Category:       s.selected.Category,
		SpeedMBps:      s.selected.Speed,
		SelectedAt:     s.selectedAt,
		Selecting:      s.selecting,
		CandidateCount: s.candidateCount,
	}
}

func chooseBestSteamCDNIP(ctx context.Context, testURL string, blocked map[string]time.Time) (steamCDNSelection, int) {
	options, err := fetchSteamCDNIPOptions(ctx)
	if err != nil || len(options) == 0 {
		options = make([]steamCDNIPOption, 0, len(defaultSteamCDNIPs))
		for _, ip := range defaultSteamCDNIPs {
			options = append(options, steamCDNIPOption{IP: ip, Category: "内置"})
		}
	}

	now := time.Now()
	available := options[:0]
	for _, option := range options {
		if until, failed := blocked[option.IP]; failed && now.Before(until) {
			continue
		}
		available = append(available, option)
	}
	if len(available) == 0 {
		return steamCDNSelection{}, len(options)
	}

	type latencyResult struct {
		option  steamCDNIPOption
		latency time.Duration
	}
	latencyCtx, latencyCancel := context.WithTimeout(ctx, 3*time.Second)
	defer latencyCancel()
	latencies := make(chan latencyResult, len(available))
	var latencyWG sync.WaitGroup
	for _, option := range available {
		latencyWG.Add(1)
		go func(option steamCDNIPOption) {
			defer latencyWG.Done()
			if latency, err := probeSteamCDNTLSLatency(latencyCtx, option.IP); err == nil {
				latencies <- latencyResult{option: option, latency: latency}
			}
		}(option)
	}
	latencyWG.Wait()
	close(latencies)

	reachable := make([]latencyResult, 0, len(available))
	for result := range latencies {
		reachable = append(reachable, result)
	}
	if len(reachable) == 0 {
		return steamCDNSelection{}, len(options)
	}
	sort.Slice(reachable, func(i, j int) bool { return reachable[i].latency < reachable[j].latency })
	if len(reachable) > steamCDNSpeedTestLimit {
		reachable = reachable[:steamCDNSpeedTestLimit]
	}

	type speedResult struct {
		option steamCDNIPOption
		speed  float64
	}
	speedCtx, speedCancel := context.WithTimeout(ctx, 15*time.Second)
	defer speedCancel()
	speeds := make(chan speedResult, len(reachable))
	var speedWG sync.WaitGroup
	for _, candidate := range reachable {
		speedWG.Add(1)
		go func(option steamCDNIPOption) {
			defer speedWG.Done()
			if speed, err := measureSteamCDNSpeed(speedCtx, option.IP, testURL); err == nil {
				speeds <- speedResult{option: option, speed: speed}
			}
		}(candidate.option)
	}
	speedWG.Wait()
	close(speeds)

	best := steamCDNSelection{}
	for result := range speeds {
		if result.speed > best.Speed {
			best = steamCDNSelection{IP: result.option.IP, Category: result.option.Category, Speed: result.speed}
		}
	}
	if best.IP != "" {
		log.Printf("Steam CDN auto-selected %s (%s) at %.2f MB/s from %d candidates", best.IP, best.Category, best.Speed, len(options))
	}
	return best, len(options)
}

func fetchSteamCDNIPOptions(ctx context.Context) ([]steamCDNIPOption, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, steamCDNIPListAPI, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam CDN IP list returned %s", response.Status)
	}
	var raw []steamCDNIPOption
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&raw); err != nil {
		return nil, err
	}
	result := make([]steamCDNIPOption, 0, len(raw))
	seen := make(map[string]bool, len(raw))
	for _, option := range raw {
		ip := strings.TrimSpace(option.IP)
		if net.ParseIP(ip) == nil || seen[ip] {
			continue
		}
		seen[ip] = true
		option.IP = ip
		option.Category = strings.TrimSpace(option.Category)
		result = append(result, option)
		if len(result) == steamCDNCandidateLimit {
			break
		}
	}
	return result, nil
}

func probeSteamCDNTLSLatency(ctx context.Context, ip string) (time.Duration, error) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 1500 * time.Millisecond, KeepAlive: 30 * time.Second},
		Config:    &tls.Config{ServerName: logic.SteamCDNHost, MinVersion: tls.VersionTLS12},
	}
	start := time.Now()
	connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, "443"))
	if err != nil {
		return 0, err
	}
	_ = connection.Close()
	return time.Since(start), nil
}

func measureSteamCDNSpeed(ctx context.Context, ip, testURL string) (float64, error) {
	parsed, err := url.Parse(testURL)
	if err != nil || !isSteamCDNDownloadHost(parsed.Hostname()) {
		return 0, fmt.Errorf("invalid Steam CDN speed test URL")
	}
	dialer := &net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	transport := newDownloadHTTPTransport(func() string { return ip }, dialer.DialContext)
	transport.TLSHandshakeTimeout = 3 * time.Second
	transport.ResponseHeaderTimeout = 5 * time.Second
	transport.DisableKeepAlives = true
	defer transport.CloseIdleConnections()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Set("Range", fmt.Sprintf("bytes=0-%d", steamCDNSpeedTestBytes-1))
	request.Header.Set("User-Agent", defaultDownloadUserAgent())
	request.Header.Set("Referer", "https://steamcommunity.com/")
	start := time.Now()
	response, err := (&http.Client{Transport: transport}).Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("Steam CDN speed test returned %s", response.Status)
	}
	written, err := io.CopyN(io.Discard, response.Body, steamCDNSpeedTestBytes)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if written == 0 {
		return 0, fmt.Errorf("Steam CDN speed test returned no data")
	}
	duration := time.Since(start)
	if duration <= 0 {
		return 0, fmt.Errorf("Steam CDN speed test duration is invalid")
	}
	return float64(written) / 1024 / 1024 / duration.Seconds(), nil
}
