package controller

import (
	"errors"
	"l4d2-manager-next/consts"
	"l4d2-manager-next/logic"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func Clear(c *gin.Context) {
	defer LogOp(c, "清空所有地图文件")()
	mutex.Lock()
	defer mutex.Unlock()

	fileBytes, err := os.ReadFile(consts.MapListFilePath)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "获取地图记录文件失败: %v", err)
		return
	}

	fileList := strings.Split(string(fileBytes), "\n")
	retainedMaps := make([]string, 0)
	stagedMaps := make([]stagedMapRemoval, 0, len(fileList))
	for _, file := range fileList {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		staged, err := stageMapRemoval(file)
		if err != nil {
			retainedMaps = append(retainedMaps, file)
			continue
		}
		stagedMaps = append(stagedMaps, staged)
	}

	if err := writeMapList(retainedMaps); err != nil {
		for index := len(stagedMaps) - 1; index >= 0; index-- {
			_ = stagedMaps[index].rollback()
		}
		FailWithError(c, http.StatusInternalServerError, "写入剩余地图记录失败: %v", err)
		return
	}
	if err := logic.RetainMapVPKInspections(retainedMaps); err != nil {
		log.Printf("更新地图 VPK 检测记录失败: %v", err)
	}
	var cleanupErrors []error
	for _, staged := range stagedMaps {
		if err := staged.cleanup(); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	if len(retainedMaps) > 0 || len(cleanupErrors) > 0 {
		FailWithError(c, http.StatusInternalServerError, "部分地图清理失败，保留：%s；暂存清理错误：%v", strings.Join(retainedMaps, ","), errors.Join(cleanupErrors...))
		return
	}

	c.String(http.StatusOK, "清空成功！")
}
