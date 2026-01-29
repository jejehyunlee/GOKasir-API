package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// Response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// Success response helper
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created response (201)
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error response helper
func Error(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// SuccessWithMeta for paginated responses
func SuccessWithMeta(c *gin.Context, message string, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ValidationError for 422 responses
func ValidationError(c *gin.Context, message string, errors interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Message: message,
		Error:   errors,
	})
}

// NotFound response
func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, resource+" not found", nil)
}

// BadRequest response
func BadRequest(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusBadRequest, message, err)
}

// InternalServerError response
func InternalServerError(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusInternalServerError, message, err)
}

// Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil)
}

func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Cek pattern error foreign key PostgreSQL atau MySQL
	return strings.Contains(errStr, "foreign key constraint") ||
		strings.Contains(errStr, "1452") || // MySQL foreign key error code
		strings.Contains(errStr, "23503") // PostgreSQL foreign key violation
}
