package logic

import (
	"l4d2-manager-next/consts"
	"os"
	"path/filepath"
	"testing"
)

func siteTestDirectory(t *testing.T) string {
	t.Helper()
	old := consts.ManagerDataPath
	consts.ManagerDataPath = t.TempDir()
	t.Cleanup(func() { consts.ManagerDataPath = old })
	return consts.ManagerDataPath
}

func TestSiteConfigDefaultsDoNotIncludeDeploymentIdentity(t *testing.T) {
	siteTestDirectory(t)
	config, err := GetSiteConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.ICPNumber != "" || config.UpdateRepository != "ledslad/l4d2-server-next" {
		t.Fatalf("unexpected defaults: %#v", config)
	}
}

func TestSiteConfigPersistsAndEnvironmentOverrides(t *testing.T) {
	dir := siteTestDirectory(t)
	_, err := SetSiteConfig(SiteConfig{ICPNumber: "测试备案号", UpdateRepository: "example/community-next"})
	if err != nil {
		t.Fatal(err)
	}
	config, err := GetSiteConfig()
	if err != nil || config.ICPNumber != "测试备案号" {
		t.Fatalf("config=%#v err=%v", config, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "site_config.json")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("L4D2_MANAGER_ICP_NUMBER", "")
	t.Setenv("L4D2_MANAGER_UPDATE_REPOSITORY", "another/fork")
	config, err = GetSiteConfig()
	if err != nil || config.ICPNumber != "" || config.UpdateRepository != "another/fork" {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}

func TestSiteConfigRejectsURLsAndInvalidFile(t *testing.T) {
	dir := siteTestDirectory(t)
	for _, value := range []string{"https://github.com/owner/repo", "owner/repo/../../private", "bad owner/repo"} {
		if _, err := SetSiteConfig(SiteConfig{UpdateRepository: value}); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "site_config.json"), []byte("not-json"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := GetSiteConfig(); err == nil {
		t.Fatal("silently accepted invalid config")
	}
}
