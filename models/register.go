package models

// RegisterRequest คือ Struct ที่รับข้อมูลสมัครสมาชิกจาก JSON Request Body
type RegisterRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Username     string `json:"username" validate:"required,username"`
	Password     string `json:"password" validate:"required,min=6,max=20"`
	LineID       string `json:"line_id" validate:"omitempty"`
	Phone        string `json:"phone" validate:"required,phone"`
	BusinessType string `json:"business_type" validate:"required"`
	WebsiteName  string `json:"website_name" validate:"required,min=2,max=30,website"`
}
