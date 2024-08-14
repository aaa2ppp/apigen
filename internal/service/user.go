package service

import (
	"context"
	"log"
)

// apigen:api {"url": "/users", "method": "POST"}
func (api *Service) CreateUser(ctx context.Context, params CreateUser) (NewUser, error) {
	const op = "CreateUser"
	// TODO
	log.Printf("%s: %v", op, params)
	return NewUser{ID: 1}, nil
}

// apigen:api {"url": "/users", "method": "GET", "auth": true}
func (api *Service) GetUser(ctx context.Context, params GetUser) (User, error) {
	const op = "GetUser"
	// TODO
	log.Printf("%s: %v", op, params)
	return User{ID: 1, Name: "Vasya", Skill: 100500, Latency: 10}, nil
}

// apigen:api {"url": "/users", "method": "PUT", "auth": true}
func (api *Service) UpdateUser(ctx context.Context, params UpdateUser) (None, error) {
	const op = "UpdateUser"
	// TODO
	log.Printf("%s: %v", op, params)
	return None{}, nil
}

// apigen:api {"url": "/users", "method": "DELETE", "auth": true}
func (api *Service) DeleteUser(ctx context.Context, params DeleteUser) (None, error) {
	const op = "DeleteUser"
	// TODO
	log.Printf("%s: %v", op, params)
	return None{}, nil
}
