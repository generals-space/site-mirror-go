package model

import "github.com/jinzhu/gorm"

// queryUnfinishedTasks ...
func queryUnfinishedTasks(db *gorm.DB, urlType int) (tasks []*URLRecord, err error) {
	tasks = []*URLRecord{}
	err = db.Where("url_type = ? and status in (?)", urlType, []int{URLTaskStatusInit, URLTaskStatusPending}).Find(&tasks).Error
	return
}

// QueryUnfinishedPageTasks ...
func QueryUnfinishedPageTasks(db *gorm.DB) (tasks []*URLRecord, err error) {
	return queryUnfinishedTasks(db, URLTypePage)
}

// QueryUnfinishedAssetTasks ...
func QueryUnfinishedAssetTasks(db *gorm.DB) (tasks []*URLRecord, err error) {
	return queryUnfinishedTasks(db, URLTypeAsset)
}

// AddURLRecord 添加URLRecord新记录(如果已存在则无操作)
func AddURLRecord(db *gorm.DB, task *URLRecord) (err error) {
	urlRecord := &URLRecord{}
	err = db.Where("url = ?", task.URL).First(urlRecord).Error
	if err != nil {
		if err.Error() != "record not found" {
			logger.Errorf("查询url记录出错: url: %s, error: %s", task.URL, err.Error())
			return
		}
	}
	err = db.Create(task).Error
	return
}

// UpdateURLRecordStatus 更新url任务记录状态
func UpdateURLRecordStatus(db *gorm.DB, url string, status int) (err error) {
	urlRecord := &URLRecord{}
	err = db.Where("url = ?", url).First(urlRecord).Error
	if err != nil {
		if err.Error() != "record not found" {
			logger.Errorf("查询url记录出错: url: %s, error: %s", url, err.Error())
		}
		return
	}

	err = db.Model(urlRecord).UpdateColumn("status", status).Error
	return
}
