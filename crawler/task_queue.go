package crawler

import "gitee.com/generals-space/site-mirror-go.git/model"

// LoadTaskQueue 初始化任务队列, 读取数据库中的`PageTask`与`AssetTask`表,
// 将其中缓存的任务加载到任务队列中
func (crawler *Crawler) LoadTaskQueue() (err error) {
	logger.Info("初始化任务队列")
	crawler.DBClientMutex.Lock()
	pageTasks, err := model.QueryUnfinishedPageTasks(crawler.DBClient)
	crawler.DBClientMutex.Unlock()
	if err != nil {
		logger.Errorf("获取页面任务失败: %s", err.Error())
		return
	}

	logger.Debugf("获取页面队列任务数量: %d", len(pageTasks))
	for _, task := range pageTasks {
		crawler.PageQueue <- task
		// crawler.EnqueuePage(task)
	}

	crawler.DBClientMutex.Lock()
	assetTasks, err := model.QueryUnfinishedAssetTasks(crawler.DBClient)
	crawler.DBClientMutex.Unlock()
	if err != nil {
		logger.Errorf("获取页面任务失败: %s", err.Error())
		return
	}

	logger.Debugf("获取静态资源队列任务数量: %d", len(pageTasks))
	for _, task := range assetTasks {
		crawler.AssetQueue <- task
		// crawler.EnqueueAsset(task)
	}
	logger.Infof("初始化任务队列完成, 页面任务数量: %d, 静态资源任务数量: %d", len(crawler.PageQueue), len(crawler.AssetQueue))
	return
}

// EnqueuePage 页面任务入队列.
// 入队列前查询数据库记录, 如已有记录则不再接受.
// 已进入队列的任务, 必定已经存在记录, 但不一定能成功下载.
// 由于队列长度有限, 这里可能会阻塞, 最可能发生死锁
// 每个page worker在解析页面时, 会将页面中的链接全部入队列.
// 如果此时队列已满, page worker就会阻塞, 当所有worker都阻塞到这里时, 程序就无法继续执行.
func (crawler *Crawler) EnqueuePage(req *model.URLRecord) {
	var err error

	crawler.PageQueue <- req

	crawler.DBClientMutex.Lock()
	defer crawler.DBClientMutex.Unlock()

	err = model.AddOrUpdateURLRecord(crawler.DBClient, req)
	if err != nil {
		logger.Errorf("添加(更新)页面任务url记录失败, req: %+v, err: %s", req, err.Error())
		return
	}
	return
}

// EnqueueAsset 页面任务入队列.
// 入队列前查询数据库记录, 如已有记录则不再接受.
func (crawler *Crawler) EnqueueAsset(req *model.URLRecord) {
	var err error
	// 由于队列长度有限, 这里可能会阻塞
	crawler.AssetQueue <- req

	crawler.DBClientMutex.Lock()
	defer crawler.DBClientMutex.Unlock()

	err = model.AddOrUpdateURLRecord(crawler.DBClient, req)
	if err != nil {
		logger.Errorf("添加(更新)静态资源任务url记录失败, req: %+v, err: %s", req, err.Error())
		return
	}
	return
}
