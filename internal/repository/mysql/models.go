package mysql

import (
	"time"

	"gorm.io/gorm"
)

type TaskModel struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"index"`
	TaskName  string
	Status    string
	Result    string
	CreatedAt time.Time
}

func (TaskModel) TableName() string {
	return "tasks"
}

type ChatMessageModel struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"index"`
	Role      string
	Content   string
	CreatedAt time.Time
}

func (ChatMessageModel) TableName() string {
	return "chat_messages"
}

type ReminderModel struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"index"`
	Title     string
	Content   string
	TriggerAt time.Time
}

func (ReminderModel) TableName() string {
	return "reminders"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&TaskModel{}, &ChatMessageModel{}, &ReminderModel{})
}
