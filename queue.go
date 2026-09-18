package main

import (
	"crypto/rand"
	"errors"
	"log"
	"time"
)

const taskTopic = "tasks.created"

var errQueueFull = errors.New("task queue is full")

type Task struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
}

// Producer publishes tasks to a queue shared with the consumer.
type Producer struct {
	queue chan<- Task
}

func (p Producer) Publish(name string) (Task, error) {
	task := Task{
		ID:        rand.Text(),
		Name:      name,
		Topic:     taskTopic,
		CreatedAt: time.Now().UTC(),
	}
	select {
	case p.queue <- task:
		return task, nil
	default:
		return Task{}, errQueueFull
	}
}

// consume processes tasks until the queue is closed and drained.
func consume(queue <-chan Task, logger *log.Logger) {
	for task := range queue {
		logger.Printf("processed task id=%s topic=%s name=%q", task.ID, task.Topic, task.Name)
	}
}
