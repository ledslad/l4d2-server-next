package controller

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const requestUrl = "http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001"
const userStatsUrl = "http://api.steampowered.com/ISteamUserStats/GetUserStatsForGame/v0002"

var client = resty.New()

func GetUserPlaytime(c *gin.Context) {
	steamID := c.PostForm("steamid")
	if steamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供SteamID"})
		return
	}

	totalPlaytime, realPlaytime, err := getPlaytimeDetails(steamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_playtime": totalPlaytime,
		"real_playtime":  realPlaytime,
	})
}

// Steam API 响应结构体 (GetOwnedGames)
type SteamGamesResponse struct {
	Response struct {
		Games []struct {
			AppID           int `json:"appid"`
			PlaytimeForever int `json:"playtime_forever"`
		} `json:"games"`
	} `json:"response"`
}

// Steam API 响应结构体 (GetUserStatsForGame)
type SteamUserStatsResponse struct {
	PlayerStats struct {
		SteamID  string `json:"steamID"`
		GameName string `json:"gameName"`
		Stats    []struct {
			Name  string  `json:"name"`
			Value float64 `json:"value"`
		} `json:"stats"`
	} `json:"playerstats"`
}

func getPlaytimeDetails(steamID string) (float64, float64, error) {
	key := os.Getenv("STEAM_API_KEY")
	if key == "" {
		return 0, 0, fmt.Errorf("未设置STEAM_API_KEY，请在https://steamcommunity.com/dev/apikey获取并设置")
	}

	id, err := convertSteamIDToSteam64ID(steamID)
	if err != nil {
		return 0, 0, fmt.Errorf("无效的SteamID: %w", err)
	}

	steam64ID := strconv.FormatUint(id, 10)

	var wg sync.WaitGroup
	var totalPlaytime, realPlaytime float64
	var errTotal, errReal error

	wg.Add(2)

	// 1. 获取总时长 (GetOwnedGames)
	go func() {
		defer wg.Done()
		gamesRes := &SteamGamesResponse{}
		_, err := client.R().SetQueryParams(map[string]string{
			"key":     key,
			"steamid": steam64ID,
		}).SetResult(gamesRes).Get(requestUrl)

		if err != nil {
			errTotal = err
			return
		}

		for _, game := range gamesRes.Response.Games {
			if game.AppID == 550 {
				totalPlaytime = float64(game.PlaytimeForever) / 60
				return
			}
		}
	}()

	// 2. 获取实战时长 (GetUserStatsForGame)
	go func() {
		defer wg.Done()
		userStatsRes := &SteamUserStatsResponse{}
		resp, err := client.R().SetQueryParams(map[string]string{
			"appid":   "550",
			"key":     key,
			"steamid": steam64ID,
		}).SetResult(userStatsRes).Get(userStatsUrl)

		if err != nil {
			errReal = err
			return
		}

		if resp.StatusCode() == 200 && userStatsRes.PlayerStats.Stats != nil {
			for _, stat := range userStatsRes.PlayerStats.Stats {
				if stat.Name == "Stat.TotalPlayTime.Total" {
					realPlaytime = stat.Value / 3600.0
					return
				}
			}
		}
	}()

	wg.Wait()

	// 如果两个都失败了，或者其中一个关键失败导致无法返回任何有意义的数据
	// 这里只要有一个成功就算成功，都失败则报错
	if totalPlaytime == 0 && realPlaytime == 0 {
		if errTotal != nil {
			return 0, 0, fmt.Errorf("获取总时长失败: %v", errTotal)
		}
		if errReal != nil {
			return 0, 0, fmt.Errorf("获取实战时长失败: %v", errReal)
		}
		return 0, 0, fmt.Errorf("未找到玩家的游戏数据，可能资料未公开")
	}

	return totalPlaytime, realPlaytime, nil
}

func convertSteamIDToSteam64ID(steamID string) (uint64, error) {
	// 使用正则表达式匹配 SteamID 的各个部分
	re := regexp.MustCompile(`^STEAM_([0-1]):([0-1]):(\d+)$`)
	matches := re.FindStringSubmatch(steamID)

	if len(matches) != 4 {
		return 0, fmt.Errorf("invalid SteamID format: %s", steamID)
	}

	// 提取 Y 和 Z 的值
	y, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid Y value in SteamID: %w", err)
	}
	z, err := strconv.ParseUint(matches[3], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid Z value in SteamID: %w", err)
	}

	// Steam64 ID 的基数
	const steam64Base uint64 = 76561197960265728

	// 计算 Steam64 ID
	steam64ID := z*2 + steam64Base + y

	return steam64ID, nil
}
