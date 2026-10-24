package logic

import (
	"encoding/json"
	"fmt"
	"l4d2-manager-next/consts"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

// DefaultUpdateRepository is the community fork, independent of upstream releases.
// A deployment can override it without rebuilding via L4D2_MANAGER_UPDATE_REPOSITORY.
var DefaultUpdateRepository = "ledslad/l4d2-server-next"

type SiteConfig struct {
	ICPNumber        string `json:"icp_number"`
	UpdateRepository string `json:"update_repository"`
}

var siteConfigMu sync.Mutex
var githubRepositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

func NormalizeSiteConfig(config SiteConfig) (SiteConfig, error) {
	config.ICPNumber = strings.TrimSpace(config.ICPNumber)
	config.UpdateRepository = strings.TrimSpace(config.UpdateRepository)
	if utf8.RuneCountInString(config.ICPNumber) > 64 || strings.ContainsAny(config.ICPNumber, "\r\n\x00") {
		return SiteConfig{}, fmt.Errorf("备案号最多 64 个字符且不能包含换行")
	}
	if config.UpdateRepository != "" && (!githubRepositoryPattern.MatchString(config.UpdateRepository) || strings.Contains(config.UpdateRepository, "..")) {
		return SiteConfig{}, fmt.Errorf("更新仓库格式应为 GitHub 用户名/仓库名")
	}
	return config, nil
}

func GetSiteConfig() (SiteConfig, error) {
	siteConfigMu.Lock()
	defer siteConfigMu.Unlock()
	config := SiteConfig{UpdateRepository: DefaultUpdateRepository}
	data, err := os.ReadFile(filepath.Join(consts.ManagerDataPath, "site_config.json"))
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return SiteConfig{}, fmt.Errorf("站点配置文件无效")
		}
	} else if !os.IsNotExist(err) {
		return SiteConfig{}, err
	}
	if value, ok := os.LookupEnv("L4D2_MANAGER_ICP_NUMBER"); ok {
		config.ICPNumber = value
	}
	if value, ok := os.LookupEnv("L4D2_MANAGER_UPDATE_REPOSITORY"); ok {
		config.UpdateRepository = value
	}
	return NormalizeSiteConfig(config)
}

func SetSiteConfig(config SiteConfig) (SiteConfig, error) {
	config, err := NormalizeSiteConfig(config)
	if err != nil {
		return SiteConfig{}, err
	}
	siteConfigMu.Lock()
	defer siteConfigMu.Unlock()
	if err := consts.EnsureManagerDataPath(); err != nil {
		return SiteConfig{}, err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return SiteConfig{}, err
	}
	temp, err := os.CreateTemp(consts.ManagerDataPath, ".site-config-*.json")
	if err != nil {
		return SiteConfig{}, err
	}
	defer os.Remove(temp.Name())
	if err = temp.Chmod(0600); err == nil {
		_, err = temp.Write(data)
	}
	if err == nil {
		err = temp.Sync()
	}
	closeErr := temp.Close()
	if err != nil {
		return SiteConfig{}, err
	}
	if closeErr != nil {
		return SiteConfig{}, closeErr
	}
	if err := os.Rename(temp.Name(), filepath.Join(consts.ManagerDataPath, "site_config.json")); err != nil {
		return SiteConfig{}, err
	}
	return config, nil
}
