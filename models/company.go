package models

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Name    string `json:"name" gorm:"size:100;not null"`
	Address string `json:"address" gorm:"size:255"`
	Phone   string `json:"phone" gorm:"size:30"`
	Email   string `json:"email" gorm:"size:254"`
}

type CompanyRequest struct {
	Name    string `json:"name" validate:"required,max=100"`
	Address string `json:"address" validate:"max=255"`
	Phone   string `json:"phone" validate:"max=30"`
	Email   string `json:"email" validate:"omitempty,email,max=254"`
}
