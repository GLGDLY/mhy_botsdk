package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v3"

	bot_apis "github.com/GLGDLY/mhy_botsdk/apis"
	bot_base "github.com/GLGDLY/mhy_botsdk/bot"
)

func renderProgess(l_limit, u_limit, progress uint64) string {
	percent := float64(progress-l_limit) / float64(u_limit-l_limit) * 100
	numBlocks := int(percent / 20) // how many blocks over 5 blocks
	if numBlocks > 5 {
		numBlocks = 5
	} else if numBlocks < 0 {
		numBlocks = 0
	}
	ret := fmt.Sprintf("%.0f%% |", percent)
	for i := 0; i < int(numBlocks); i++ {
		ret += "█"
	}
	for i := 0; i < 5-int(numBlocks); i++ {
		ret += " "
	}
	ret += fmt.Sprintf("| (%d/%d)", progress-l_limit, u_limit-l_limit)
	return ret
}

type Config struct {
	BotID     string `yaml:"bot_id"`
	BotSecret string `yaml:"bot_secret"`
	PublicKey string `yaml:"public_key"`
	Range     uint64 `yaml:"range"`
}

func (c *Config) LoadConfig() {
	f, err := os.Open("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("无法打开配置文件 config.yaml，请检查是否存在该文件: %v", err))
	}
	defer f.Close()
	yaml.NewDecoder(f).Decode(c)
}

var mu sync.Mutex

func getAllVilla(api *bot_apis.ApiBase, start_i, end_i uint64, all_villa *[]interface{}, progress *uint64) {
	var i uint64
outer_for:
	for i = start_i; i < end_i; i++ {
		*progress++
		resp, http, err := api.GetVilla(i)
		for {
			if http != 200 || err != nil { // mainly handle 429 too many requests
				time.Sleep(2 * time.Millisecond)
				resp, http, err = api.GetVilla(i)
				continue
			}
			if resp.Data.Villa.VillaID == 0 {
				continue outer_for
			}
			break
		}
		mu.Lock()
		*all_villa = append(*all_villa, resp.Data.Villa)
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
}

func main() {
	var config Config
	config.LoadConfig()
	if config.BotID == "" || config.BotSecret == "" || config.PublicKey == "" {
		fmt.Println("config.yaml invalid")
		return
	}
	if config.Range == 0 {
		config.Range = 3000
	}
	fmt.Println("正在获取从 0 到", config.Range, "的所有villa信息...")
	var bot = bot_base.NewBot(config.BotID, config.BotSecret, config.PublicKey, "/", ":8888")

	var all_villa []interface{} = make([]interface{}, 0)

	limit := [3]uint64{config.Range / 3, config.Range / 3 * 2, config.Range}
	progress := limit
	for i := 0; i < 3; i++ {
		progress[i] -= config.Range / 3
	}

	for i := 0; i < 3; i++ {
		go getAllVilla(bot.Api, limit[i]-config.Range/3, limit[i], &all_villa, &progress[i])
	}

	fmt.Println("||==================================||==================================||==================================||")

	for progress[0]+progress[1]+progress[2] < config.Range*2 {
		// \033[2K\r
		out := [3]string{}
		for i := 0; i < 3; i++ {
			out[i] = fmt.Sprintf("[t_%d]: %s", i+1, renderProgess(limit[i]-config.Range/3, limit[i], progress[i]))
		}
		fmt.Printf("\r|| %-32s || %-32s || %-32s ||", out[0], out[1], out[2])
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("\n已加入别野数量：", len(all_villa), "，正在写入文件...")

	f, _ := os.Create("all_villa.json")
	defer f.Close()
	json.NewEncoder(f).Encode(all_villa)

	fmt.Println("写入 \"all_villa.json\" 完成，按回车键退出")
	fmt.Scanln()
}
