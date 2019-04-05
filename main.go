package main

import (
	"os"
	"os/signal"
	"syscall"

	"gitee.com/generals-space/site-mirror-go.git/crawler"
	"gitee.com/generals-space/site-mirror-go.git/util"
)

var logger = util.NewLogger(os.Stdout)

func main() {
	config := &crawler.Config{
		PageQueueSize:    5000,
		AssetQueueSize:   50,
		PageWorkerCount:  1,
		AssetWorkerCount: 1,
		SiteDBPath:       "site.db",
		SitePath:         "sites",
		StartPage:        "https://www.lewenxiaoshuo.com/",
		MaxDepth:         2,
	}

	c, err := crawler.NewCrawler(config)
	if err != nil {
		panic(err)
	}
	c.Start()

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	logger.Info(<-channel)
}
