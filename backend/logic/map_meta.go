package logic

import (
	"encoding/json"
	"fmt"
	"l4d2-manager-next/consts"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WorkshopMeta is stored beside a VPK and is compatible with LYTVPK metadata.
type WorkshopMeta struct {
	WorkshopID   string `json:"workshop_id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Description  string `json:"description"`
	PreviewURL   string `json:"preview_url"`
	FileURL      string `json:"file_url"`
	DownloadedAt string `json:"downloaded_at"`
	TimeUpdated  string `json:"time_updated"`
}

func MapMetaPath(vpkPath string) string {
	return strings.TrimSuffix(vpkPath, filepath.Ext(vpkPath)) + ".meta"
}

func LoadMapWorkshopMeta(mapName string) (*WorkshopMeta, error) {
	mapName, err := NormalizeMapVPKName(mapName)
	if err != nil {
		return nil, err
	}
	return LoadWorkshopMetaForVPK(filepath.Join(consts.AddonsBasePath, mapName))
}

func LoadWorkshopMetaForVPK(vpkPath string) (*WorkshopMeta, error) {
	data, err := os.ReadFile(MapMetaPath(vpkPath))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var meta WorkshopMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse map metadata: %w", err)
	}
	meta.WorkshopID = strings.TrimSpace(meta.WorkshopID)
	if !isValidWorkshopID(meta.WorkshopID) {
		return nil, fmt.Errorf("invalid workshop ID in map metadata")
	}
	return &meta, nil
}

func SaveWorkshopMetaForVPK(vpkPath string, meta WorkshopMeta) error {
	meta.WorkshopID = strings.TrimSpace(meta.WorkshopID)
	if !isValidWorkshopID(meta.WorkshopID) {
		return fmt.Errorf("invalid workshop ID")
	}
	if strings.TrimSpace(meta.DownloadedAt) == "" {
		meta.DownloadedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	metaPath := MapMetaPath(vpkPath)
	temp, err := os.CreateTemp(filepath.Dir(metaPath), ".map-meta-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, metaPath)
}
