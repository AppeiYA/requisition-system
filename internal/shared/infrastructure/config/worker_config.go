package config

import "time"

// WorkerConfig configures the asynq worker process and the tasks enqueued
// for it.
type WorkerConfig struct {
	Concurrency     int
	Queue           string
	TaskMaxRetry    int
	TaskTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// LoadWorkerConfig loads the background worker settings. Defaults keep a
// single-queue setup running out of the box.
func LoadWorkerConfig() (*WorkerConfig, error) {
	concurrency, err := getIntEnv("WORKER_CONCURRENCY", 10)
	if err != nil {
		return nil, err
	}

	taskMaxRetry, err := getIntEnv("WORKER_TASK_MAX_RETRY", 5)
	if err != nil {
		return nil, err
	}

	taskTimeout, err := getTimeDurationEnv("WORKER_TASK_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}

	shutdownTimeout, err := getTimeDurationEnv("WORKER_SHUTDOWN_TIMEOUT", 8*time.Second)
	if err != nil {
		return nil, err
	}

	return &WorkerConfig{
		Concurrency:     concurrency,
		Queue:           getEnv("WORKER_QUEUE", "default"),
		TaskMaxRetry:    taskMaxRetry,
		TaskTimeout:     taskTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}