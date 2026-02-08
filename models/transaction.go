package models

import (
	"time"
)

type Transaction struct {
	ID          uint                `json:"id" gorm:"primaryKey"`
	TotalAmount int                 `json:"total_amount" gorm:"not null"`
	CreatedAt   time.Time           `json:"created_at"`
	Details     []TransactionDetail `json:"details" gorm:"foreignKey:TransactionID"`
}

type TransactionDetail struct {
	ID            uint     `json:"id" gorm:"primaryKey"`
	TransactionID uint     `json:"transaction_id" gorm:"not null;index"`
	ProductID     uint     `json:"product_id" gorm:"not null;index"`
	ProductName   string   `json:"product_name,omitempty" gorm:"size:100"`
	Quantity      int      `json:"quantity" gorm:"not null"`
	Subtotal      int      `json:"subtotal" gorm:"not null"`
	Product       *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

type CheckoutItem struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type CheckoutRequest struct {
	Items []CheckoutItem `json:"items"`
}
