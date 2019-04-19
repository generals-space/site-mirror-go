package model

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // 注释防止绿色下划线语法提示

	"gitee.com/generals-space/site-mirror-go.git/util"
)

var logger = util.NewLogger(os.Stdout)

const (
	// URLTaskStatusInit 任务状态初始值, 0
	URLTaskStatusInit = iota
	// URLTaskStatusPending 从队列中取出, 未出结果时的状态.
	URLTaskStatusPending
	// URLTaskStatusSuccess 任务状态成功, 2
	URLTaskStatusSuccess
	// URLTaskStatusFailed 任务状态失败(404), 3
	URLTaskStatusFailed
)

const (
	URLTypePage int = iota
	URLTypeAsset
)

// URLRecord 任务记录表
type URLRecord struct {
	gorm.Model
	URL         string `gorm:"unique, not null"`
	Refer       string
	Depth       int
	URLType     int
	FailedTimes int
	Status      int `gorm:"default 0"`
}

// GetDB 获取数据库链接
func GetDB(dbPath string) (db *gorm.DB, err error) {
	db, err = gorm.Open("sqlite3", dbPath)
	tables := []interface{}{
		&URLRecord{},
	}
	db.AutoMigrate(tables...)
	return
}
