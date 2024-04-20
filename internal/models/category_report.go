package models

type CategoryReport struct {
	Name       string
	Amount     int64
	Percentage float64 `json:"percentage"`
}
