package mysql

import (
	"context"

	reminderdomain "campus-agent/internal/domain/reminder"
	"gorm.io/gorm"
)

type ReminderRepository struct {
	db *gorm.DB
}

func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

func (r *ReminderRepository) Save(ctx context.Context, reminder reminderdomain.Reminder) error {
	model := ReminderModel{
		ID:        reminder.ID,
		UserID:    reminder.UserID,
		Title:     reminder.Title,
		Content:   reminder.Content,
		TriggerAt: reminder.TriggerAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *ReminderRepository) FindByID(ctx context.Context, id int64) (reminderdomain.Reminder, error) {
	var model ReminderModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return reminderdomain.Reminder{}, err
	}

	return reminderdomain.Reminder{
		ID:        model.ID,
		UserID:    model.UserID,
		Title:     model.Title,
		Content:   model.Content,
		TriggerAt: model.TriggerAt,
	}, nil
}
