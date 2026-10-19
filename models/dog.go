package models

import "gorm.io/gorm"

type Dogs struct {
	gorm.Model
	Name  string `json:"name" gorm:"size:100;not null"`
	DogID int    `json:"dog_id" gorm:"not null"`
}

// แยกข้อมูลรับเข้า เพื่อไม่ให้ Client แก้ ID หรือ DeletedAt ของ GORM
type DogRequest struct {
	Name  string `json:"name" validate:"required,max=100"`
	DogID *int   `json:"dog_id" validate:"required,gte=0"`
}
