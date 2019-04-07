package model

import "github.com/jinzhu/gorm"

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

// DelAssetTask 删除静态资源任务
func DelAssetTask(db *gorm.DB, req *URLTask) (err error) {
	err = db.Unscoped().Where("url = ?", req.URL).Delete(&AssetTask{}).Error
	if err != nil {
		logger.Errorf("删除静态资源任务队列记录失败: req: %+v, error: %s", req, err.Error())
		return
	}
	logger.Debugf("已从数据库移除静态资源队列任务对象: %+v", req)
	return
}
