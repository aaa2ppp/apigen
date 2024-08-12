package api

import (
	"context"

	"matchmaker/internal/model"
)

type ApiError = model.ApiError

type UserCreator interface {
	Create(ctx context.Context, params model.CreateUser) (model.NewUser, error)
}

type Api struct {
	UserCreator
}

// apigen:api {"url": "/users", "method": "POST"}
func (api *Api) CreateUser(ctx context.Context, params CreateUserParams) (NewUser, error) {
	return api.UserCreator.Create(ctx, model.CreateUser(params))
}

type CreateUserParams struct {
	Name    string  `apivalidator:"required"`
	Skill   float64 `apivalidator:"required,>=0"`
	Latency float64 `apivalidator:"required,>0"`
}

type NewUser = model.NewUser
