package mysql

import (
	"context"

	taskdomain "campus-agent/internal/domain/task"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Save(ctx context.Context, task taskdomain.Task) error {
	model := taskToModel(task)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id int64, status taskdomain.Status, result string) error {
	return r.db.WithContext(ctx).
		Model(&TaskModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status": string(status),
			"result": result,
		}).Error
}

func (r *TaskRepository) FindByID(ctx context.Context, id int64) (taskdomain.Task, error) {
	var model TaskModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return taskdomain.Task{}, err
	}
	return taskFromModel(model), nil
}

func (r *TaskRepository) ListByUser(ctx context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	if limit <= 0 {
		limit = 20
	}

	var models []TaskModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Order("id desc").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}

	tasks := make([]taskdomain.Task, 0, len(models))
	for _, model := range models {
		tasks = append(tasks, taskFromModel(model))
	}
	return tasks, nil
}

func taskToModel(task taskdomain.Task) TaskModel {
	return TaskModel{
		ID:        task.ID,
		UserID:    task.UserID,
		TaskName:  task.TaskName,
		Status:    string(task.Status),
		Result:    task.Result,
		CreatedAt: task.CreatedAt,
	}
}

func taskFromModel(model TaskModel) taskdomain.Task {
	return taskdomain.Task{
		ID:        model.ID,
		UserID:    model.UserID,
		TaskName:  model.TaskName,
		Status:    taskdomain.Status(model.Status),
		Result:    model.Result,
		CreatedAt: model.CreatedAt,
	}
}
