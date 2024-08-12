package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"matchmaker/internal/api"
	"matchmaker/internal/model"
)

const (
	readTimeout  = time.Duration(time.Second * 60)
	writeTimeout = time.Duration(time.Second * 60)
)

func main() {
	// mux := http.NewServeMux()
	// mux.Handle("/", handler.NotFound)
	// mux.Handle("/ok", handler.Ok)
	// mux.Handle("/users", handler.NewCreateUser(userCreatorStub{})) // TODO

	srv := http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      &api.Api{UserCreator: &userCreatorStub{}},
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

type userCreatorStub struct{}

func (c userCreatorStub) Create(ctx context.Context, user model.CreateUser) (model.NewUser, error) {
	log.Printf("Create user: %+v", user)
	return model.NewUser{ID: 1}, nil
}
