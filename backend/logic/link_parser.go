package logic

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	LinkSourceWorkshop        = "workshop"
	LinkSourceQQFlashTransfer = "qq_flash_transfer"
)

type LinkParseItem struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Filename       string `json:"filename"`
	FileSize       string `json:"file_size"`
	FileURL        string `json:"file_url"`
	PreviewURL     string `json:"preview_url"`
	Author         string `json:"author"`
	Description    string `json:"description"`
	TimeUpdated    string `json:"time_updated"`
	Referer        string `json:"referer"`
	Supported      bool   `json:"supported"`
	DisabledReason string `json:"disabled_reason"`
}

type LinkParseResult struct {
	SourceType string          `json:"source_type"`
	SourceID   string          `json:"source_id"`
	Items      []LinkParseItem `json:"items"`
}

func ParseDownloadLink(rawLink string) (LinkParseResult, error) {
	rawLink = strings.TrimSpace(rawLink)
	if rawLink == "" {
		return LinkParseResult{}, fmt.Errorf("链接不能为空")
	}

	if IsQQFlashTransferLink(rawLink) {
		return ParseQQFlashTransferLink(rawLink)
	}

	if _, err := ParseWorkshopID(rawLink); err == nil {
		return ParseWorkshopDownloadLinkAsGeneric(rawLink)
	}

	return LinkParseResult{}, fmt.Errorf("暂不支持该链接类型")
}

func ParseWorkshopDownloadLinkAsGeneric(workshopURL string) (LinkParseResult, error) {
	result, err := ParseWorkshopDownloadLink(workshopURL)
	if err != nil {
		return LinkParseResult{}, err
	}

	items := make([]LinkParseItem, 0, len(result.Items))
	for _, item := range result.Items {
		filename := cleanWorkshopDownloadFilename(item)
		items = append(items, LinkParseItem{
			ID:          item.PublishedFileID,
			Title:       item.Title,
			Filename:    filename,
			FileSize:    item.FileSize,
			FileURL:     item.FileURL,
			PreviewURL:  item.PreviewURL,
			Author:      item.Creator,
			Description: item.Description,
			TimeUpdated: item.TimeUpdated,
			Supported:   true,
		})
	}

	return LinkParseResult{
		SourceType: LinkSourceWorkshop,
		SourceID:   result.SourceID,
		Items:      items,
	}, nil
}

func IsQQFlashTransferLink(rawLink string) bool {
	u, err := url.Parse(strings.TrimSpace(rawLink))
	if err != nil || u.Scheme == "" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	if host != "qfile.qq.com" && !strings.HasSuffix(host, ".qfile.qq.com") {
		return false
	}

	return strings.HasPrefix(strings.TrimRight(u.EscapedPath(), "/"), "/q/")
}

func IsSupportedDownloadFilename(filename string) bool {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(filename))) {
	case ".vpk", ".zip", ".rar", ".7z":
		return true
	default:
		return false
	}
}

func downloadItemSupportState(filename string) (bool, string) {
	if IsSupportedDownloadFilename(filename) {
		return true, ""
	}
	return false, "仅支持 .vpk, .zip, .rar, .7z 文件"
}

func cleanWorkshopDownloadFilename(item WorkshopDownloadItem) string {
	if filename := WorkshopVPKFilename(item.PublishedFileID); filename != "" {
		return filename
	}
	filename := strings.TrimSpace(item.Filename)
	if filename == "" {
		if publishedFileID := strings.TrimSpace(item.PublishedFileID); publishedFileID != "" {
			filename = publishedFileID + ".vpk"
		} else {
			filename = "workshop.vpk"
		}
	}

	if ext := strings.ToLower(filepath.Ext(filename)); ext == "" {
		filename += ".vpk"
	}

	return cleanLinkParseFilename(filename)
}

func cleanLinkParseFilename(filename string) string {
	filename = strings.TrimSpace(strings.ReplaceAll(filename, "\x00", ""))
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = filepath.Base(filename)
	filename = strings.Trim(filename, " .")

	if filename == "" || filename == "." || filename == "/" {
		return ""
	}

	return filename
}
