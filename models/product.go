package models

import "time"

type Product struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	Price      float64   `json:"price" gorm:"not null"`
	Stock      int       `json:"stock" gorm:"not null"`
	CategoryID uint      `json:"-" gorm:"not null;index"` // <-- tambah json:"-"
	Category   Category  `json:"category" gorm:"constraint:OnUpdate:NO ACTION,OnDelete:SET NULL;"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
