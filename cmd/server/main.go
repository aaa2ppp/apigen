package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"matchmaker/internal/service"
)

const (
	serverAddr   = "0.0.0.0:8080"
	readTimeout  = time.Duration(60 * time.Second)
	writeTimeout = time.Duration(60 * time.Second)
)

func main() {
	srv := http.Server{
		Addr:         serverAddr,
		Handler:      &service.Service{},
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	log.Printf("HTTP server starts on http://%v", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-done
}
