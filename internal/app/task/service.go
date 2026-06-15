package taskapp

import (
	"context"
	"errors"

	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/message"
)

type Producer interface {
	Publish(ctx context.Context, msg message.TaskMessage) error
}

type Service struct {
	repo     taskdomain.Repository
	producer Producer
}

func NewService(repo taskdomain.Repository, producer Producer) *Service {
	return &Service{
		repo:     repo,
		producer: producer,
	}
}

func (s *Service) CreateAsyncTask(ctx context.Context, task taskdomain.Task) error {
	if s.producer == nil {
		return errors.New("task producer is not configured")
	}

	if s.repo != nil {
		if err := s.repo.Save(ctx, task); err != nil {
			return err
		}
	}

	return s.producer.Publish(ctx, message.TaskMessage{
		TaskID:   task.ID,
		UserID:   task.UserID,
		TaskName: task.TaskName,
	})
}

func (s *Service) GetTask(ctx context.Context, id int64) (taskdomain.Task, error) {
	if s.repo == nil {
		return taskdomain.Task{}, errors.New("task repository is not configured")
	}

	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListTasks(ctx context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	if s.repo == nil {
		return nil, errors.New("task repository is not configured")
	}

	return s.repo.ListByUser(ctx, userID, limit)
}
