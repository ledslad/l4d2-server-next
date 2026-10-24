package controller

import (
	"l4d2-manager-next/logic"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Remove(c *gin.Context) {
	mapName, err := logic.NormalizeMapVPKName(c.PostForm("map"))
	if err != nil {
		FailWithError(c, http.StatusBadRequest, "地图名称无效")
		return
	}
	defer LogOp(c, "删除地图文件: "+mapName)()

	mutex.Lock()
	defer mutex.Unlock()

	mapList, err := readMapList()
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "删除时读取maplist.txt失败: %v", err)
		return
	}
	staged, err := stageMapRemoval(mapName)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "删除地图失败: %v", err)
		return
	}
	if err := writeMapList(removeMapNamesFromList(mapList, []string{mapName})); err != nil {
		_ = staged.rollback()
		FailWithError(c, http.StatusInternalServerError, "删除时写入文件失败: %v", err)
		return
	}
	if err := staged.cleanup(); err != nil {
		FailWithError(c, http.StatusInternalServerError, "地图已从列表移除，但清理暂存文件失败: %v", err)
		return
	}
	if err := logic.DeleteMapVPKInspection(mapName); err != nil {
		FailWithError(c, http.StatusInternalServerError, "地图已删除，但清理检测记录失败: %v", err)
		return
	}

	c.String(http.StatusOK, "删除成功！")
}
