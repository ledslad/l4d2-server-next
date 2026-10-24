package controller

import (
	"l4d2-manager-next/logic"
	"net/http"

	"github.com/gin-gonic/gin"
)

type updateDownloadConfigRequest struct {
	SteamCDNIP string `json:"steam_cdn_ip"`
}

func requireDownloadConfigAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}

	FailWithError(c, http.StatusForbidden, "需要管理员权限")
	return false
}

func GetDownloadConfig(c *gin.Context) {
	if !requireDownloadConfigAdmin(c) {
		return
	}

	response := gin.H{
		"steam_cdn_ip": logic.GetSteamCDNIP(),
	}
	snapshot := globalSteamCDNSelector.Snapshot()
	response["steam_cdn_mode"] = snapshot.Mode
	response["steam_cdn_active_ip"] = snapshot.ActiveIP
	response["steam_cdn_category"] = snapshot.Category
	response["steam_cdn_speed_mbps"] = snapshot.SpeedMBps
	response["steam_cdn_selected_at"] = snapshot.SelectedAt
	response["steam_cdn_selecting"] = snapshot.Selecting
	response["steam_cdn_candidate_count"] = snapshot.CandidateCount
	c.JSON(http.StatusOK, response)
}

func SetDownloadConfig(c *gin.Context) {
	if !requireDownloadConfigAdmin(c) {
		return
	}

	var req updateDownloadConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		FailWithError(c, http.StatusBadRequest, "参数错误: %v", err)
		return
	}

	normalizedIP, err := logic.NormalizeSteamCDNIP(req.SteamCDNIP)
	if err != nil {
		FailWithError(c, http.StatusBadRequest, "%v", err)
		return
	}

	detail := "清除 Steam CDN 指定 IP"
	if normalizedIP != "" {
		detail = "设置 Steam CDN 指定 IP: " + normalizedIP
	}
	defer LogOp(c, detail)()

	savedIP, err := logic.SetSteamCDNIP(normalizedIP)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "保存下载配置失败: %v", err)
		return
	}

	closeDownloadIdleConnections()
	globalSteamCDNSelector.Reset()
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"steam_cdn_ip": savedIP,
	})
}
