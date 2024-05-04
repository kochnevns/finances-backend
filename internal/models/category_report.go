package models

type CategoryReport struct {
	Name       string
	Color      string `json:"color"`
	Amount     int64
	Percentage float64 `json:"percentage"`
}
