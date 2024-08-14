package service

type ApiError struct {
	HTTPStatus int
	Err        error
}

func (ae ApiError) Error() string {
	return ae.Err.Error()
}

type None struct{}

type CreateUser struct {
	Name    string  `json:"name" apivalidator:"required"`
	Skill   float64 `json:"skill" apivalidator:"default=0,>=0"`
	Latency float64 `json:"latency" apivalidator:"default=1,>0"`
}

type UpdateUser struct {
	ID      int     `json:"id" apivalidator:"required,>0"`
	Name    string  `json:"name" apivalidator:"required"`
	Skill   float64 `json:"skill" apivalidator:"required,>=0"`
	Latency float64 `json:"latency" apivalidator:"required,>0"`
}

type DeleteUser struct {
	ID int `json:"id" apivalidator:"required,>0"`
}

type GetUser struct {
	ID int `json:"id" apivalidator:"required,>0"`
}

type NewUser struct {
	ID int `json:"id"`
}

type User struct {
	ID      int     `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	Skill   float64 `json:"skill,omitempty"`
	Latency float64 `json:"latency,omitempty"`
}
