package model

import "time"

// Article 示例模型。纯数据结构，不含任何方法。
type Article struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Tags      []string  `gorm:"type:text[];default:'{}'" json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
