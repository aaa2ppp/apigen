package model

type CreateUser struct {
	Name    string  `json:"name,omitempty"`
	Skill   float64 `json:"skill,omitempty"`
	Latency float64 `json:"latency,omitempty"`
}

type User struct {
	ID      uint    `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	Skill   float64 `json:"skill,omitempty"`
	Latency float64 `json:"latency,omitempty"`
}

type NewUser struct {
	ID uint `json:"id,omitempty"`
}
