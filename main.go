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
		PageQueueSize:  10,
		AssetQueueSize: 10,
		SiteDBPath:     "site.db",
		StartPage:      "http://www.abx.la/",
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
