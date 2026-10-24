package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"l4d2-manager-next/consts"
	"l4d2-manager-next/logic"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func GetSiteConfig(c *gin.Context) {
	config, err := logic.GetSiteConfig()
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "读取站点配置失败")
		return
	}
	c.JSON(http.StatusOK, config)
}

func SetSiteConfig(c *gin.Context) {
	if !requireDownloadConfigAdmin(c) {
		return
	}
	var request logic.SiteConfig
	if err := c.ShouldBindJSON(&request); err != nil {
		FailWithError(c, http.StatusBadRequest, "参数错误")
		return
	}
	if _, err := logic.NormalizeSiteConfig(request); err != nil {
		FailWithError(c, http.StatusBadRequest, "%v", err)
		return
	}
	defer LogOp(c, "更新站点备案与版本检查仓库配置")()
	if _, err := logic.SetSiteConfig(request); err != nil {
		FailWithError(c, http.StatusInternalServerError, "保存站点配置失败")
		return
	}
	config, err := logic.GetSiteConfig()
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "读取生效配置失败")
		return
	}
	c.JSON(http.StatusOK, config)
}

type latestRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

var latestReleaseMu sync.Mutex
var latestReleaseCache struct {
	Repository string
	Expires    time.Time
	Release    latestRelease
}

func fetchLatestRelease(ctx context.Context, client *http.Client, baseURL, repository string) (latestRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/repos/"+repository+"/releases/latest", nil)
	if err != nil {
		return latestRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "l4d2-server-next-community")
	response, err := client.Do(req)
	if err != nil {
		return latestRelease{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return latestRelease{}, fmt.Errorf("GitHub 返回 HTTP %d", response.StatusCode)
	}
	var release latestRelease
	if err := json.NewDecoder(io.LimitReader(response.Body, 1024*1024)).Decode(&release); err != nil {
		return latestRelease{}, err
	}
	if release.TagName == "" || !strings.HasPrefix(release.HTMLURL, "https://github.com/"+repository+"/releases/") {
		return latestRelease{}, fmt.Errorf("无效的版本信息")
	}
	return release, nil
}

func GetLatestVersion(c *gin.Context) {
	config, err := logic.GetSiteConfig()
	if err != nil || config.UpdateRepository == "" {
		FailWithError(c, http.StatusServiceUnavailable, "尚未配置本分支的 GitHub 更新仓库")
		return
	}
	latestReleaseMu.Lock()
	cached := latestReleaseCache
	latestReleaseMu.Unlock()
	release := cached.Release
	if cached.Repository != config.UpdateRepository || time.Now().After(cached.Expires) {
		client := &http.Client{Timeout: 15 * time.Second}
		release, err = fetchLatestRelease(c.Request.Context(), client, "https://api.github.com", config.UpdateRepository)
		if err != nil {
			FailWithError(c, http.StatusBadGateway, "暂时无法检查本分支更新: %v", err)
			return
		}
		latestReleaseMu.Lock()
		latestReleaseCache.Repository, latestReleaseCache.Release, latestReleaseCache.Expires = config.UpdateRepository, release, time.Now().Add(5*time.Minute)
		latestReleaseMu.Unlock()
	}
	c.JSON(http.StatusOK, gin.H{"current_version": consts.Version, "latest": release, "repository": config.UpdateRepository})
}
