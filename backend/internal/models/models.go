package models

import "gorm.io/gorm"

type APIKey struct {
	gorm.Model
	Provider   string  `json:"provider" gorm:"column:provider"`
	Name       string  `json:"name" gorm:"column:name"`
	Value      string  `json:"value" gorm:"column:value"`
	Status     string  `json:"status" gorm:"column:status"`
	BalanceUSD float64 `json:"balance_usd" gorm:"column:balance_usd"`
	CostModel  string  `json:"cost_model" gorm:"column:cost_model"`
}

type Project struct {
	gorm.Model
	Name   string `json:"name" gorm:"column:name"`
	APIKey string `json:"api_key" gorm:"column:api_key"`
	Status string `json:"status" gorm:"column:status"`
}

type Task struct {
	gorm.Model
	ProjectID   uint     `json:"project_id" gorm:"column:project_id"`
	Name        string   `json:"name" gorm:"column:name"`
	Description string   `json:"description" gorm:"column:description"`
	APIMethod   string   `json:"api_method" gorm:"column:api_method"`
	Version     string   `json:"version" gorm:"column:version"`
	Status      string   `json:"status" gorm:"column:status"`
	Prompts     []Prompt `gorm:"foreignKey:TaskID" json:"prompts"`
}

type Prompt struct {
	gorm.Model
	TaskID uint   `json:"task_id" gorm:"column:task_id"`
	Order  int    `json:"order" gorm:"column:order"`
	Name   string `json:"name" gorm:"column:name"`
	Models string `json:"model" gorm:"column:model"`
	Text   string `json:"text" gorm:"column:text"`
}

type PromptHistory struct {
	gorm.Model
	PromptID  uint    `json:"prompt_id" gorm:"column:prompt_id"`
	TaskID    uint    `json:"task_id" gorm:"column:task_id"`
	ProjectID uint    `json:"project_id" gorm:"column:project_id"`
	Input     string  `json:"input" gorm:"column:input"`
	Output    string  `json:"output" gorm:"column:output"`
	Models    string  `json:"model" gorm:"column:model"`
	Tokens    int     `json:"tokens" gorm:"column:tokens"`
	CostUSD   float64 `json:"cost_usd" gorm:"column:cost_usd"`
}
