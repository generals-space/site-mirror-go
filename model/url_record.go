package model

import "github.com/jinzhu/gorm"

// IsExistInURLRecord 查询数据库中指定的url任务记录
func IsExistInURLRecord(db *gorm.DB, url string) (exist bool) {
	var err error
	urlRecord := &URLRecord{}
	err = db.Where("url = ?", url).First(urlRecord).Error
	if err != nil {
		if err.Error() != "record not found" {
			logger.Errorf("查询url记录出错: url: %s, error: %s", url, err.Error())
		}
		exist = false
		return
	}
	exist = true
	return
}

// AddURLRecord 添加URLRecord新记录
func AddURLRecord(db *gorm.DB, task *URLTask) (err error) {
	urlRecordModel := &URLRecord{
		URLTask: *task,
	}
	err = db.Create(urlRecordModel).Error
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
