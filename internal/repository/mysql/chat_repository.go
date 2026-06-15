package mysql

import (
	"context"

	chatdomain "campus-agent/internal/domain/chat"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Save(ctx context.Context, message chatdomain.Message) error {
	model := ChatMessageModel{
		ID:        message.ID,
		UserID:    message.UserID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *ChatRepository) ListByUser(ctx context.Context, userID int64, limit int) ([]chatdomain.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	var models []ChatMessageModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Order("id desc").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}

	messages := make([]chatdomain.Message, 0, len(models))
	for _, model := range models {
		messages = append(messages, chatdomain.Message{
			ID:        model.ID,
			UserID:    model.UserID,
			Role:      model.Role,
			Content:   model.Content,
			CreatedAt: model.CreatedAt,
		})
	}
	return messages, nil
}
