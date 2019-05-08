package model

import "github.com/jinzhu/gorm"

// isExistInURLRecord 查询数据库中指定的url任务记录
func isExistInURLRecord(db *gorm.DB, url string) (exist bool) {
	var err error
	urlRecord := &URLRecord{}
	err = db.Where("url = ?", url).First(urlRecord).Error
	if err != nil {
		exist = false
		return
	}
	exist = true
	return
}

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

// AddOrUpdateURLRecord 任务入队列时添加URLRecord新记录(如果已存在则更新failed_times和status字段)
func AddOrUpdateURLRecord(db *gorm.DB, task *URLRecord) (err error) {
	exist := isExistInURLRecord(db, task.URL)
	if exist {
		whereArgs := map[string]interface{}{
			"url": task.URL,
		}
		dataToBeUpdated := map[string]interface{}{
			"failed_times": task.FailedTimes,
			"status":       URLTaskStatusInit, // 任务重新入队列要将状态修改为init状态
		}
		err = db.Model(&URLRecord{}).Where(whereArgs).Updates(dataToBeUpdated).Error
	} else {
		err = db.Create(task).Error
	}
	return
}

// UpdateURLRecordStatus 更新url任务记录状态
func UpdateURLRecordStatus(db *gorm.DB, url string, status int) (err error) {
	urlRecord := &URLRecord{}
	err = db.Where("url = ?", url).First(urlRecord).Error
	if err != nil {
		return
	}

	err = db.Model(urlRecord).UpdateColumn("status", status).Error
	return
}
