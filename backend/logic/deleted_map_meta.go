package logic

import (
	"encoding/json"
	"fmt"
	"l4d2-manager-next/consts"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DeletedMapRemoval contains the map information that is still available
// before the VPK and its sidecar metadata are removed.
type DeletedMapRemoval struct {
	Filename            string
	ID                  string
	Size                string
	WorkshopID          string
	WorkshopIDs         []string
	Titles              []string
	Maps                []string
	ManagerSummaryError string
	ManagerDetailError  string
}

// NewDeletedMapRemoval snapshots the metadata needed by the deleted-map
// manifest. Parsing failures are retained in the record and never block a
// successful file deletion.
func NewDeletedMapRemoval(mapName string) DeletedMapRemoval {
	removal := DeletedMapRemoval{
		Filename:    mapName,
		ID:          strings.TrimSuffix(mapName, filepath.Ext(mapName)),
		Size:        "unknown",
		WorkshopIDs: []string{},
		Titles:      []string{},
		Maps:        []string{},
	}

	vpkPath := filepath.Join(consts.AddonsBasePath, mapName)
	if info, err := os.Stat(vpkPath); err == nil && !info.IsDir() {
		removal.Size = formatDeletedMapSize(info.Size())
	}

	if meta, err := LoadWorkshopMetaForVPK(vpkPath); err == nil && meta != nil {
		removal.WorkshopID = meta.WorkshopID
		removal.WorkshopIDs = []string{meta.WorkshopID}
	}

	campaigns, err := GetMapMissionDetail(mapName)
	if err != nil {
		removal.ManagerSummaryError = fmt.Sprintf("解析地图名称失败: %v", err)
		return removal
	}
	for _, campaign := range campaigns {
		if campaign == nil {
			continue
		}
		if title := strings.TrimSpace(campaign.Title); title != "" && !containsDeletedMapValue(removal.Titles, title) {
			removal.Titles = append(removal.Titles, title)
		}
		for _, chapter := range campaign.Chapters {
			if chapter == nil {
				continue
			}
			if code := strings.TrimSpace(chapter.Code); code != "" && !containsDeletedMapValue(removal.Maps, code) {
				removal.Maps = append(removal.Maps, code)
			}
		}
	}
	return removal
}

// PersistDeletedMapRemoval writes the same manifest consumed by l4d2-service.
// An unfinished service deletion record is updated in place so one deletion
// cannot produce duplicate records.
func (removal DeletedMapRemoval) PersistDeletedMapRemoval(status, managerResponse string) error {
	if removal.Filename == "" {
		return fmt.Errorf("deleted map filename is empty")
	}
	metaDir := filepath.Join(consts.AddonsBasePath, "deleted-vpk-meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return err
	}

	var targetPath string
	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(entries))
	prefix := strings.ToLower(removal.Filename + ".deleted-")
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, prefix) && strings.HasSuffix(lowerName, ".json") {
			paths = append(paths, filepath.Join(metaDir, name))
		}
	}
	sort.Strings(paths)
	for index := len(paths) - 1; index >= 0; index-- {
		data, readErr := os.ReadFile(paths[index])
		if readErr != nil {
			continue
		}
		var existing map[string]any
		if json.Unmarshal(data, &existing) != nil {
			continue
		}
		existingStatus, _ := existing["status"].(string)
		existingFilename, _ := existing["filename"].(string)
		if strings.EqualFold(existingFilename, removal.Filename) && existingStatus != "removed" {
			targetPath = paths[index]
			break
		}
	}

	stamp := time.Now().Format("20060102-150405-000000")
	record := map[string]any{
		"removed_at":            stamp,
		"query":                 removal.ID,
		"id":                    removal.ID,
		"workshop_id":           nil,
		"workshop_ids":          removal.WorkshopIDs,
		"filename":              removal.Filename,
		"size":                  removal.Size,
		"titles":                removal.Titles,
		"maps":                  removal.Maps,
		"manager_summary_error": removal.ManagerSummaryError,
		"manager_detail_error":  removal.ManagerDetailError,
		"status":                status,
		"manager_response":      managerResponse,
	}
	if removal.WorkshopID != "" {
		record["workshop_id"] = removal.WorkshopID
	}

	if targetPath != "" {
		data, readErr := os.ReadFile(targetPath)
		if readErr != nil {
			return readErr
		}
		var existing map[string]any
		if err := json.Unmarshal(data, &existing); err != nil {
			return err
		}
		if existingRemovedAt, ok := existing["removed_at"].(string); ok && existingRemovedAt != "" {
			record["removed_at"] = existingRemovedAt
		}
		// Keep the service's query, associations and parser diagnostics. The
		// Manager only completes the pending record with its final status.
		existing["status"] = status
		existing["manager_response"] = managerResponse
		if _, ok := existing["manager_summary_error"]; !ok || existing["manager_summary_error"] == "" {
			existing["manager_summary_error"] = removal.ManagerSummaryError
		}
		if _, ok := existing["manager_detail_error"]; !ok || existing["manager_detail_error"] == "" {
			existing["manager_detail_error"] = removal.ManagerDetailError
		}
		record = existing
	} else {
		targetPath = filepath.Join(metaDir, removal.Filename+".deleted-"+stamp+".json")
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temp, err := os.CreateTemp(metaDir, ".deleted-vpk-meta-*.tmp")
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
	return os.Rename(tempPath, targetPath)
}

func containsDeletedMapValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func formatDeletedMapSize(size int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case size >= gb:
		return fmt.Sprintf("%.2fGB", float64(size)/gb)
	case size >= mb:
		return fmt.Sprintf("%.2fMB", float64(size)/mb)
	case size >= kb:
		return fmt.Sprintf("%.2fKB", float64(size)/kb)
	default:
		return fmt.Sprintf("%dB", size)
	}
}
