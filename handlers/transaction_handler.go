package handlers

import (
	"Kasir-API/models"
	"Kasir-API/services"
	"Kasir-API/utils"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) Checkout(c *gin.Context) {
	var request models.CheckoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ValidationError(c, "Invalid request payload", err.Error())
		return
	}

	transaction, err := h.service.Checkout(request)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	utils.Created(c, "Transaction completed successfully", transaction)
}

func (h *TransactionHandler) GetAll(c *gin.Context) {
	transactions, err := h.service.GetAll()
	if err != nil {
		utils.InternalServerError(c, "Failed to retrieve transactions", err.Error())
		return
	}

	utils.Success(c, "Transactions retrieved successfully", transactions)
}
