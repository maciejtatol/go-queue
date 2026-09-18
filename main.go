package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Manually wire one producer and one consumer to a bounded, in-memory queue.
	queue := make(chan Task, 100)
	producer := Producer{queue: queue}
	consumerDone := make(chan struct{})
	go func() {
		consume(queue, log.Default())
		close(consumerDone)
	}()

	server := &http.Server{
		Addr:              ":8080",
		Handler:           newHandler(producer),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		log.Println("HTTP server listening on http://localhost:8080")
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server failed: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown failed: %v", err)
		return
	}
	// All producers have finished, so it is safe to close and drain the queue.
	close(queue)
	<-consumerDone
}
