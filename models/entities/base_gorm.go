package entities

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 自定义基础模型，使用 uint64 作为 ID
type BaseModel struct {
	ID        uint64         `gorm:"primarykey"` // 定义 ID 为 uint64 且是主键
	CreatedAt time.Time      // 创建时间
	UpdatedAt time.Time      // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除时间戳，并添加索引
}
