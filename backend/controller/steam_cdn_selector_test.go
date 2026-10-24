package controller

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSteamCDNSelectorCachesAndReplacesFailedIP(t *testing.T) {
	setupDownloadConfigControllerTest(t)
	selector := newSteamCDNSelector()
	now := time.Unix(1_700_000_000, 0)
	selector.now = func() time.Time { return now }
	selections := []steamCDNSelection{
		{IP: "192.0.2.10", Category: "first", Speed: 10},
		{IP: "192.0.2.20", Category: "second", Speed: 8},
	}
	chooseCalls := 0
	selector.choose = func(context.Context, string, map[string]time.Time) (steamCDNSelection, int) {
		selection := selections[chooseCalls]
		chooseCalls++
		return selection, 2
	}

	const testURL = "https://cdn.steamusercontent.com/ugc/test/file"
	if got := selector.Select(context.Background(), testURL); got != "192.0.2.10" {
		t.Fatalf("first selected IP = %q", got)
	}
	if got := selector.Select(context.Background(), testURL); got != "192.0.2.10" || chooseCalls != 1 {
		t.Fatalf("cached selected IP = %q, choose calls = %d", got, chooseCalls)
	}

	selector.MarkFailed("192.0.2.10")
	if got := selector.Select(context.Background(), testURL); got != "192.0.2.20" || chooseCalls != 2 {
		t.Fatalf("replacement selected IP = %q, choose calls = %d", got, chooseCalls)
	}
}

func TestSteamCDNSelectorSharesConcurrentSelection(t *testing.T) {
	setupDownloadConfigControllerTest(t)
	selector := newSteamCDNSelector()
	started := make(chan struct{})
	release := make(chan struct{})
	chooseCalls := 0
	selector.choose = func(context.Context, string, map[string]time.Time) (steamCDNSelection, int) {
		chooseCalls++
		close(started)
		<-release
		return steamCDNSelection{IP: "192.0.2.30", Speed: 5}, 1
	}

	const testURL = "https://cdn.steamusercontent.com/ugc/test/file"
	results := make(chan string, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- selector.Select(context.Background(), testURL)
		}()
	}
	<-started
	close(release)
	wg.Wait()
	close(results)
	for result := range results {
		if result != "192.0.2.30" {
			t.Fatalf("concurrent selected IP = %q", result)
		}
	}
	if chooseCalls != 1 {
		t.Fatalf("choose calls = %d, want 1", chooseCalls)
	}
}

func TestDownloadTaskExposesPreferredFilenameBeforeConnection(t *testing.T) {
	semaphore := make(chan struct{})
	task := newDownloadTask("https://example.invalid/file", " expected-map.vpk ", "", nil, semaphore)
	task.Cancel()
	if got := task.GetFilename(); got != "expected-map.vpk" {
		t.Fatalf("task filename = %q, want expected-map.vpk", got)
	}
}
