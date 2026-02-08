package services

import (
	"Kasir-API/database"
	"Kasir-API/models"
	"Kasir-API/repositories"
	"fmt"
)

type TransactionService struct {
	repo *repositories.TransactionRepository
}

func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Checkout(request models.CheckoutRequest) (*models.Transaction, error) {
	if len(request.Items) == 0 {
		return nil, fmt.Errorf("checkout items cannot be empty")
	}

	var transaction models.Transaction
	var totalAmount int
	var details []models.TransactionDetail

	db := database.GetDB()

	for _, item := range request.Items {
		var product models.Product
		if err := db.First(&product, item.ProductID).Error; err != nil {
			return nil, fmt.Errorf("product ID %d not found", item.ProductID)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product: %s", product.Name)
		}

		subtotal := int(product.Price) * item.Quantity
		totalAmount += subtotal

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	transaction.TotalAmount = totalAmount
	transaction.Details = details

	if err := s.repo.Create(&transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (s *TransactionService) GetAll() ([]models.Transaction, error) {
	return s.repo.GetAll()
}

func (s *TransactionService) GetReport(startDate, endDate string) (models.ReportResponse, error) {
	return s.repo.GetReport(startDate, endDate)
}
