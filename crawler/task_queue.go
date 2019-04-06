package crawler

import "gitee.com/generals-space/site-mirror-go.git/model"

// LoadTaskQueue 初始化任务队列, 读取数据库中的`PageTask`与`AssetTask`表,
// 将其中缓存的任务加载到任务队列中
func (crawler *Crawler) LoadTaskQueue() (err error) {
	logger.Debug("初始化任务队列")
	pageTasks, err := model.QueryPageTasks(crawler.DBClient)
	if err != nil {
		logger.Errorf("获取页面任务失败: %s", err.Error())
		return
	}
	for _, task := range pageTasks {
		crawler.EnqueuePage(&task.URLTask)
	}
	assetTasks, err := model.QueryAssetTasks(crawler.DBClient)
	if err != nil {
		logger.Errorf("获取页面任务失败: %s", err.Error())
		return
	}
	for _, task := range assetTasks {
		crawler.EnqueuePage(&task.URLTask)
	}
	logger.Debug("初始化任务队列完成")
	return
}

// EnqueuePage 页面任务入队列.
// 入队列前查询数据库记录, 如已有记录则不再接受.
// 已进入队列的任务, 必定已经存在记录, 但不一定能成功下载.
func (crawler *Crawler) EnqueuePage(req *model.URLTask) {
	var err error
	exist := model.IsExistInURLRecord(crawler.DBClient, req.URL)
	if exist {
		return
	}
	// 由于队列长度有限, 这里可能会阻塞.
	crawler.PageQueue <- req
	tx := crawler.DBClient.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	err = model.AddURLRecord(tx, req)
	if err != nil {
		logger.Errorf("添加页面任务url记录失败, req: %+v, err: %s", req, err.Error())
		return
	}
	err = model.AddPageTask(tx, req)
	if err != nil {
		logger.Errorf("添加页面任务队列记录失败, req: %+v, err: %s", req, err.Error())
		return
	}
	return
}

// EnqueueAsset 页面任务入队列.
// 入队列前查询数据库记录, 如已有记录则不再接受.
func (crawler *Crawler) EnqueueAsset(req *model.URLTask) {
	var err error
	exist := model.IsExistInURLRecord(crawler.DBClient, req.URL)
	if exist {
		return
	}
	// 由于队列长度有限, 这里可能会阻塞
	crawler.AssetQueue <- req
	tx := crawler.DBClient.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	err = model.AddURLRecord(tx, req)
	if err != nil {
		logger.Errorf("添加静态资源任务url记录失败, req: %+v, err: %s", req, err.Error())
		return
	}
	err = model.AddAssetTask(tx, req)
	if err != nil {
		logger.Errorf("添加静态资源任务队列记录失败, req: %+v, err: %s", req, err.Error())
		return
	}
	return
}
