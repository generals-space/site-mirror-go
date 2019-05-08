package main

import (
	"os"
	"os/signal"
	"syscall"

	"gitee.com/generals-space/site-mirror-go.git/crawler"
	"gitee.com/generals-space/site-mirror-go.git/util"
)

func main() {
	logger := util.NewLogger(os.Stdout)
	// logger.SetLevel("debug")

	config := crawler.NewConfig()
	config.StartPage = "https://www.lewenxiaoshuo.com/"
	config.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36"
	config.MaxDepth = 1

	c, err := crawler.NewCrawler(config, logger)
	if err != nil {
		panic(err)
	}
	c.Start()
	defer func() {
		logger.Info("用户取消")
	}()
	// 等待用户取消, 目前无法自动结束.
	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	logger.Info(<-channel)
}
