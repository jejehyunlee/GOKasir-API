package models

import "time"

type Product struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	Price      float64   `json:"price" gorm:"not null"`
	Stock      int       `json:"stock" gorm:"not null"`
	CategoryID *uint     `json:"category_id" gorm:"not null;index"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
