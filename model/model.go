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
	// URLTaskStatusSuccess 任务状态成功, 1
	URLTaskStatusSuccess
	// URLTaskStatusFailed 任务状态失败(404), 2
	URLTaskStatusFailed
)

// URLTask 任务对象
type URLTask struct {
	URL         string `gorm:"unique, not null"`
	Refer       string
	Depth       int
	URLType     int
	FailedTimes int
}

// URLRecord 任务记录表
type URLRecord struct {
	gorm.Model
	URLTask
	Status int `gorm:"default 0"`
}

// PageTask 页面任务队列表, 作为备份, 防止丢失.
type PageTask struct {
	gorm.Model
	URLTask
}

// AssetTask 静态资源任务队列表, 作为备份, 防止丢失.
type AssetTask struct {
	gorm.Model
	URLTask
}

// GetDB 获取数据库链接
func GetDB(dbPath string) (db *gorm.DB, err error) {
	db, err = gorm.Open("sqlite3", dbPath)
	tables := []interface{}{
		&URLRecord{}, &PageTask{}, &AssetTask{},
	}
	db.AutoMigrate(tables...)
	return
}
