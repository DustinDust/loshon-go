package app

type Response[T any] struct {
	Data  T   `json:"data"`
	Page  int `json:"page,omitempty"`
	Total int `json:"total,omitempty"`
}
