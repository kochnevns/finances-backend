package models

type Expense struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Amount      int64  `json:"amount"`
	Date        string `json:"date"`
	Category    string `json:"category"`
	CategoryID  int64  `json:"category_id"`
}
