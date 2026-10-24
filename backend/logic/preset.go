package logic

import (
	"fmt"
	"l4d2-manager-next/consts"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type PresetConfig struct {
	Platform map[string]string `yaml:"platform"`
	Presets  []Preset          `yaml:"preset"`
}

type Preset struct {
	Name    string         `yaml:"name"`
	Desc    string         `yaml:"desc"`
	Plugins []PresetPlugin `yaml:"plugins"`
}

type PresetPlugin struct {
	Name    string               `yaml:"name"`
	Configs []PresetPluginConfig `yaml:"configs"`
}

type PresetPluginConfig struct {
	Name   string            `yaml:"name"`
	Values map[string]string `yaml:"values"`
}

type PresetInfo struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	PluginCount int    `json:"plugin_count"`
}

func GetPresets() ([]PresetInfo, error) {
	data, err := os.ReadFile("preset.yaml")
	if err != nil {
		return nil, err
	}

	var config PresetConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	var presets []PresetInfo
	for _, p := range config.Presets {
		presets = append(presets, PresetInfo{
			Name:        p.Name,
			Desc:        p.Desc,
			PluginCount: len(p.Plugins),
		})
	}

	return presets, nil
}

func ApplyPreset(presetName string) error {
	data, err := os.ReadFile("preset.yaml")
	if err != nil {
		return err
	}

	var config PresetConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Determine platform plugin
	platformKey := runtime.GOOS
	if platformKey != "windows" && platformKey != "linux" {
		return fmt.Errorf("不支持的平台: %s", platformKey)
	}
	platformPlugin, ok := config.Platform[platformKey]
	if !ok {
		return fmt.Errorf("未配置平台插件: %s", platformKey)
	}

	var targetPreset *Preset
	for _, p := range config.Presets {
		if p.Name == presetName {
			targetPreset = &p
			break
		}
	}

	if targetPreset == nil {
		return fmt.Errorf("未找到预设: %s", presetName)
	}

	// 1. Validate all plugins exist
	storePath := getStorePath()
	if _, err := os.Stat(filepath.Join(storePath, platformPlugin)); os.IsNotExist(err) {
		return fmt.Errorf("未找到平台插件: %s", platformPlugin)
	}

	for _, p := range targetPreset.Plugins {
		// Skip platform plugin check here as we check it separately, but it's fine to check again
		if _, err := os.Stat(filepath.Join(storePath, p.Name)); os.IsNotExist(err) {
			return fmt.Errorf("未找到插件: %s", p.Name)
		}
	}

	// 2. Disable all currently enabled plugins
	// We need to fetch currently enabled plugins first
	// GetPlugins() returns all plugins with status.
	allPlugins, err := GetPlugins()
	if err != nil {
		return err
	}

	var toDisable []string
	for _, p := range allPlugins {
		if p.Status == "enabled" {
			toDisable = append(toDisable, p.Name)
		}
	}

	if len(toDisable) > 0 {
		if err := DisablePlugins(toDisable); err != nil {
			return fmt.Errorf("禁用当前插件失败: %v", err)
		}
	}

	// 3. Enable plugins
	// Enable platform plugin first
	if err := EnablePlugin(platformPlugin); err != nil {
		// If this fails, we are in a bad state (all plugins disabled).
		return fmt.Errorf("启用平台插件 %s 失败: %v", platformPlugin, err)
	}

	// Enable other plugins
	for _, p := range targetPreset.Plugins {
		if p.Name == platformPlugin {
			continue
		}
		if err := EnablePlugin(p.Name); err != nil {
			return fmt.Errorf("启用插件 %s 失败: %v", p.Name, err)
		}
	}

	// 4. Apply configs
	for _, p := range targetPreset.Plugins {
		for _, cfg := range p.Configs {
			// Apply config
			// Configs are usually in left4dead2/cfg/sourcemod/ or left4dead2/cfg/
			// The logic in PluginConfig assumes cfg/sourcemod for now based on SavePluginConfig
			// But cfg.Name might contain path separators if user specified relative path?
			// The yaml example shows "survivor_chat_select.cfg"

			// We should try to find where the config file should be.
			// Usually sourcemod configs are in cfg/sourcemod.
			// Let's assume cfg/sourcemod for now.

			cfgPath := filepath.Join(consts.GamePath, "cfg", "sourcemod", cfg.Name)

			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
				fmt.Printf("Warning: failed to create directory for config %s: %v\n", cfg.Name, err)
				continue
			}

			if err := UpdateOrCreateSourceModConfig(cfgPath, cfg.Values); err != nil {
				fmt.Printf("Warning: failed to apply config %s: %v\n", cfg.Name, err)
			}
		}
	}

	// Apply global preset configs if any? The yaml structure has configs under plugins.
	// But wait, the yaml has:
	// - name: ...
	//   plugins:
	//     - name: ...
	//       configs: ...
	// So configs are per plugin.

	return nil
}
