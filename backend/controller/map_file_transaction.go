package controller

import (
	"errors"
	"fmt"
	"l4d2-manager-next/consts"
	"l4d2-manager-next/logic"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type stagedMapRemoval struct {
	mapName       string
	vpkPath       string
	vpkStagePath  string
	metaPath      string
	metaStagePath string
	vpkStaged     bool
	metaStaged    bool
}

func stageMapRemoval(mapName string) (stagedMapRemoval, error) {
	normalizedName, err := logic.NormalizeMapVPKName(mapName)
	if err != nil {
		return stagedMapRemoval{}, err
	}
	token := uuid.NewString()
	vpkPath := filepath.Join(consts.AddonsBasePath, normalizedName)
	metaPath := logic.MapMetaPath(vpkPath)
	staged := stagedMapRemoval{
		mapName:       normalizedName,
		vpkPath:       vpkPath,
		vpkStagePath:  filepath.Join(consts.AddonsBasePath, ".map-remove-"+token+".vpk.tmp"),
		metaPath:      metaPath,
		metaStagePath: filepath.Join(consts.AddonsBasePath, ".map-remove-"+token+".meta.tmp"),
	}
	if moved, err := renameIfExists(staged.vpkPath, staged.vpkStagePath); err != nil {
		return stagedMapRemoval{}, fmt.Errorf("暂存地图文件失败: %w", err)
	} else {
		staged.vpkStaged = moved
	}
	if moved, err := renameIfExists(staged.metaPath, staged.metaStagePath); err != nil {
		_ = staged.rollback()
		return stagedMapRemoval{}, fmt.Errorf("暂存地图元数据失败: %w", err)
	} else {
		staged.metaStaged = moved
	}
	return staged, nil
}

func renameIfExists(source, destination string) (bool, error) {
	if err := os.Rename(source, destination); err != nil && !os.IsNotExist(err) {
		return false, err
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (staged stagedMapRemoval) rollback() error {
	var rollbackErrors []error
	if staged.metaStaged {
		if err := os.Rename(staged.metaStagePath, staged.metaPath); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	if staged.vpkStaged {
		if err := os.Rename(staged.vpkStagePath, staged.vpkPath); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	return errors.Join(rollbackErrors...)
}

func (staged stagedMapRemoval) cleanup() error {
	var cleanupErrors []error
	for _, path := range []string{staged.vpkStagePath, staged.metaStagePath} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	return errors.Join(cleanupErrors...)
}

func removeMapNamesFromList(mapList []string, names []string) []string {
	removed := make(map[string]bool, len(names))
	for _, name := range names {
		removed[name] = true
	}
	result := make([]string, 0, len(mapList))
	for _, name := range mapList {
		if !removed[strings.TrimSpace(name)] && strings.TrimSpace(name) != "" {
			result = append(result, strings.TrimSpace(name))
		}
	}
	return result
}

func rollbackInstalledMaps(mapNames []string) error {
	mutex.Lock()
	defer mutex.Unlock()

	mapList, err := readMapList()
	if err != nil {
		return err
	}
	staged := make([]stagedMapRemoval, 0, len(mapNames))
	for _, name := range mapNames {
		item, err := stageMapRemoval(name)
		if err != nil {
			for index := len(staged) - 1; index >= 0; index-- {
				_ = staged[index].rollback()
			}
			return err
		}
		staged = append(staged, item)
	}
	if err := writeMapList(removeMapNamesFromList(mapList, mapNames)); err != nil {
		for index := len(staged) - 1; index >= 0; index-- {
			_ = staged[index].rollback()
		}
		return err
	}
	var cleanupErrors []error
	for _, item := range staged {
		if err := item.cleanup(); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
		if err := logic.DeleteMapVPKInspection(item.mapName); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	return errors.Join(cleanupErrors...)
}
