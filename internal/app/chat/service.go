package chatapp

import (
	"context"

	"campus-agent/internal/agent/executor"
	"campus-agent/internal/agent/planner"
)

type RAG interface {
	Answer(ctx context.Context, query string) (string, error)
}

type Service struct {
	planner  planner.Planner
	executor executor.Executor
	rag      RAG
}

type Request struct {
	UserID  int64
	Message string
}

type Response struct {
	Intent  string                `json:"intent"`
	Tasks   []string              `json:"tasks"`
	Answer  string                `json:"answer,omitempty"`
	Results []executor.TaskResult `json:"results,omitempty"`
}

func NewService(plannerAgent planner.Planner, executorAgent executor.Executor, ragAgent RAG) *Service {
	return &Service{
		planner:  plannerAgent,
		executor: executorAgent,
		rag:      ragAgent,
	}
}

func (s *Service) Handle(ctx context.Context, req Request) (Response, error) {
	planResult, err := s.planner.Plan(ctx, req.Message)
	if err != nil {
		return Response{}, err
	}

	resp := Response{
		Intent: string(planResult.Intent),
		Tasks:  make([]string, 0, len(planResult.Tasks)),
	}

	for _, task := range planResult.Tasks {
		resp.Tasks = append(resp.Tasks, string(task.Name))
	}

	if planResult.Intent == planner.IntentKnowledgeQuery && s.rag != nil {
		answer, err := s.rag.Answer(ctx, req.Message)
		if err != nil {
			return Response{}, err
		}
		resp.Answer = answer
		return resp, nil
	}

	if len(planResult.Tasks) == 0 || s.executor == nil {
		return resp, nil
	}

	tasks := make([]planner.Task, 0, len(planResult.Tasks))
	for _, task := range planResult.Tasks {
		task.UserID = req.UserID
		task.Input = req.Message
		tasks = append(tasks, task)
	}

	results, err := s.executor.Execute(ctx, tasks)
	if err != nil {
		return Response{}, err
	}
	resp.Results = results
	return resp, nil
}
