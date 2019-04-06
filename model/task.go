package model

import "github.com/jinzhu/gorm"

// QueryPageTasks 查询页面任务队列记录, 同时删除所有记录(之后要加载到任务队列中).
func QueryPageTasks(db *gorm.DB) (tasks []*PageTask, err error) {
	tasks = []*PageTask{}
	err = db.Find(&tasks).Error
	if err != nil {
		if err.Error() == "record not found" {
			err = nil
		}
	}
	err = db.Unscoped().Delete(&PageTask{}).Error
	if err != nil {
		logger.Errorf("清空页面任务队列记录失败: %s", err.Error())
	}
	return
}

// QueryAssetTasks 查询静态资源任务队列记录, 同时删除所有记录(之后要加载到任务队列中).
func QueryAssetTasks(db *gorm.DB) (tasks []*AssetTask, err error) {
	tasks = []*AssetTask{}
	err = db.Find(&tasks).Error
	if err != nil {
		if err.Error() == "record not found" {
			err = nil
		}
	}
	err = db.Unscoped().Delete(&AssetTask{}).Error
	if err != nil {
		logger.Errorf("清空静态资源任务队列记录失败: %s", err.Error())
	}
	return
}

// AddPageTask 添加页面任务到数据库
func AddPageTask(db *gorm.DB, req *URLTask) (err error) {
	pageTaskModel := &PageTask{
		URLTask: *req,
	}
	err = db.Create(pageTaskModel).Error
	if err != nil {
		logger.Errorf("添加页面任务队列失败: req: %+v, error: %s", req, err.Error())
		return
	}
	return
}

// AddAssetTask 添加静态资源任务到数据库
func AddAssetTask(db *gorm.DB, req *URLTask) (err error) {
	assetTaskModel := &AssetTask{
		URLTask: *req,
	}
	err = db.Create(assetTaskModel).Error
	if err != nil {
		logger.Errorf("添加静态资源任务队列失败: req: %+v, error: %s", req, err.Error())
		return
	}
	return
}

// DelPageTask 删除指定任务
func DelPageTask(db *gorm.DB, req *URLTask) (err error) {
	err = db.Unscoped().Where("url = ?", req.URL).Delete(&PageTask{}).Error
	if err != nil {
		logger.Errorf("删除页面任务队列记录失败: req: %+v, error: %s", req, err.Error())
		return
	}
	return
}

// DelAssetTask 删除静态资源任务
func DelAssetTask(db *gorm.DB, req *URLTask) (err error) {
	err = db.Unscoped().Where("url = ?", req.URL).Delete(&AssetTask{}).Error
	if err != nil {
		logger.Errorf("删除静态资源任务队列记录失败: req: %+v, error: %s", req, err.Error())
		return
	}
	return
}
