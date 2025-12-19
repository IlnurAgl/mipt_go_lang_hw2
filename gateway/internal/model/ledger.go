package model

type BudgetAdd struct {
	Category string  `json:"category" example:"Продукты"`
	Limit    float64 `json:"limit" example:"1250.50"`
}

type BudgetGet struct {
	Category string `json:"category" example:"Продукты"`
}

type BudgetGetResponse struct {
	Category string  `json:"category" example:"Продукты"`
	Limit    float64 `json:"limit" example:"1250.50"`
}

type TrasnactionAdd struct {
	Amount      float64 `json:"amount" example:"123.50"`
	Category    string  `json:"category" example:"Продукты"`
	Description string  `json:"description" example:"Тест"`
	Date        string  `json:"Date" example:"2025-12-19"`
}

type TransactionAddResponse struct {
	Id int64 `json:"id" example:"1"`
}

type TransactionGet struct {
	Id int64 `json:"id" example:"1"`
}

type TransactionGetResponse struct {
	Id          int64   `json:"id" example:"1"`
	Amount      float64 `json:"amount" example:"123.50"`
	Category    string  `json:"category" example:"Продукты"`
	Description string  `json:"description" example:"Тест"`
	Date        string  `json:"Date" example:"2025-12-19"`
}
