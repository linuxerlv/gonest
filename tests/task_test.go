package tests

import (
	"context"
	"testing"
	"time"

	"github.com/linuxerlv/gonest/task"
)

func TestTaskOptions_WithMaxRetry(t *testing.T) {
	opts := &task.TaskOptions{}
	option := task.WithMaxRetry(5)

	option(opts)

	if opts.MaxRetry != 5 {
		t.Errorf("Expected MaxRetry 5, got %d", opts.MaxRetry)
	}
}

func TestTaskOptions_WithTimeout(t *testing.T) {
	opts := &task.TaskOptions{}
	option := task.WithTimeout(30 * time.Second)

	option(opts)

	if opts.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", opts.Timeout)
	}
}

func TestTaskOptions_WithUnique(t *testing.T) {
	opts := &task.TaskOptions{}
	option := task.WithUnique()

	option(opts)

	if !opts.Unique {
		t.Error("Expected Unique to be true")
	}
}

func TestQueueTask_Fields(t *testing.T) {
	now := time.Now()
	taskItem := &task.QueueTask{
		ID:        "task-123",
		Type:      "email",
		Payload:   []byte("test"),
		Priority:  1,
		Queue:     "default",
		Metadata:  map[string]string{"key": "value"},
		CreatedAt: now,
	}

	if taskItem.ID != "task-123" {
		t.Errorf("Expected ID 'task-123', got '%s'", taskItem.ID)
	}

	if taskItem.Type != "email" {
		t.Errorf("Expected Type 'email', got '%s'", taskItem.Type)
	}

	if taskItem.Priority != 1 {
		t.Errorf("Expected Priority 1, got %d", taskItem.Priority)
	}

	if taskItem.Queue != "default" {
		t.Errorf("Expected Queue 'default', got '%s'", taskItem.Queue)
	}
}

func TestJobInfo_Fields(t *testing.T) {
	now := time.Now()
	info := task.JobInfo{
		Name:       "cleanup",
		Schedule:   "0 0 * * *",
		NextRun:    now.Add(24 * time.Hour),
		LastRun:    now,
		Running:    false,
		RunCount:   10,
		ErrorCount: 0,
	}

	if info.Name != "cleanup" {
		t.Errorf("Expected Name 'cleanup', got '%s'", info.Name)
	}

	if info.Schedule != "0 0 * * *" {
		t.Errorf("Expected Schedule '0 0 * * *', got '%s'", info.Schedule)
	}

	if info.RunCount != 10 {
		t.Errorf("Expected RunCount 10, got %d", info.RunCount)
	}
}

func TestQueueStats_Fields(t *testing.T) {
	stats := task.QueueStats{
		Name:      "default",
		Pending:   100,
		Active:    10,
		Scheduled: 50,
		Retry:     5,
		Processed: 1000,
		Failed:    3,
	}

	if stats.Name != "default" {
		t.Errorf("Expected Name 'default', got '%s'", stats.Name)
	}

	if stats.Pending != 100 {
		t.Errorf("Expected Pending 100, got %d", stats.Pending)
	}

	if stats.Processed != 1000 {
		t.Errorf("Expected Processed 1000, got %d", stats.Processed)
	}
}

func TestQueueConfig_Fields(t *testing.T) {
	cfg := task.QueueConfig{
		Workers:     4,
		MaxRetry:    3,
		Concurrency: 10,
		Timeout:     30 * time.Second,
		Queues:      map[string]int{"default": 1, "high": 2},
	}

	if cfg.Workers != 4 {
		t.Errorf("Expected Workers 4, got %d", cfg.Workers)
	}

	if cfg.MaxRetry != 3 {
		t.Errorf("Expected MaxRetry 3, got %d", cfg.MaxRetry)
	}

	if len(cfg.Queues) != 2 {
		t.Errorf("Expected 2 queues, got %d", len(cfg.Queues))
	}
}

func TestSchedulerConfig_Fields(t *testing.T) {
	cfg := task.SchedulerConfig{
		Timezone:    "UTC",
		Concurrency: 5,
		Timeout:     60 * time.Second,
	}

	if cfg.Timezone != "UTC" {
		t.Errorf("Expected Timezone 'UTC', got '%s'", cfg.Timezone)
	}

	if cfg.Concurrency != 5 {
		t.Errorf("Expected Concurrency 5, got %d", cfg.Concurrency)
	}
}

func TestTaskResult_Fields(t *testing.T) {
	now := time.Now()
	result := task.TaskResult{
		TaskID: "task-123",
		Status: "completed",
		Result: []byte("success"),
		Error:  "",
		At:     now,
	}

	if result.TaskID != "task-123" {
		t.Errorf("Expected TaskID 'task-123', got '%s'", result.TaskID)
	}

	if result.Status != "completed" {
		t.Errorf("Expected Status 'completed', got '%s'", result.Status)
	}
}

func TestJobHandler_Type(t *testing.T) {
	var handler task.JobHandler = func(ctx context.Context) error {
		return nil
	}

	if handler == nil {
		t.Error("Expected handler to be defined")
	}
}

func TestTaskHandler_Type(t *testing.T) {
	var handler task.TaskHandler = func(ctx context.Context, taskItem *task.QueueTask) error {
		return nil
	}

	if handler == nil {
		t.Error("Expected handler to be defined")
	}
}
