package crawler

import "gitee.com/generals-space/site-mirror-go.git/model"

// LoadTaskQueue 初始化任务队列, 读取数据库中的`PageTask`与`AssetTask`表,
// 将其中缓存的任务加载到任务队列中
func (crawler *Crawler) LoadTaskQueue() (err error) {
	logger.Info("初始化任务队列")
	pageTasks, err := model.QueryPageTasks(crawler.DBClient)
	if err != nil {
		logger.Errorf("获取页面任务失败: %s", err.Error())
		return
	}
	logger.Debugf("获取页面队列任务数量: %d", len(pageTasks))
	for _, task := range pageTasks {
		crawler.EnqueuePage(&task.URLTask)
	}
	assetTasks, err := model.QueryAssetTasks(crawler.DBClient)
	if err != nil {
		logger.Errorf("获取页面任务失败: %s", err.Error())
		return
	}
	logger.Debugf("获取静态资源队列任务数量: %d", len(pageTasks))
	for _, task := range assetTasks {
		crawler.EnqueuePage(&task.URLTask)
	}
	logger.Infof("初始化任务队列完成, 页面任务数量: %d, 静态资源任务数量: %d", len(crawler.PageQueue), len(crawler.AssetQueue))
	return
}

// SaveTaskQueue 手动停止进程时, 将队列中的任务持久化到数据库.
func (crawler *Crawler) SaveTaskQueue() (err error) {
	logger.Infof("开始存储任务队列, 待存储页面队列任务数量: %d, 静态资源队列任务数量: %d", len(crawler.PageQueue), len(crawler.AssetQueue))
	var pageTaskSum, assetTaskSum int
	if len(crawler.PageQueue) != 0 {
		for req := range crawler.PageQueue {
			pageTaskSum++
			err = model.AddPageTask(crawler.DBClient, req)
			if err != nil {
				logger.Errorf("存储页面任务失败: %s", err.Error())
			}
			if len(crawler.PageQueue) == 0 {
				break
			}
		}
	}
	if len(crawler.PageQueue) != 0 {
		for req := range crawler.AssetQueue {
			assetTaskSum++
			err = model.AddAssetTask(crawler.DBClient, req)
			if err != nil {
				logger.Errorf("存储静态资源任务失败: %s", err.Error())
			}
			if len(crawler.PageQueue) == 0 {
				break
			}
		}
	}
	logger.Debug("队列清空, 准备关闭")
	// 关闭channel
	close(crawler.PageQueue)
	close(crawler.AssetQueue)
	logger.Infof("存储任务队列完成, 页面任务数量: %d, 静态资源任务数量: %d", pageTaskSum, assetTaskSum)
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
