package middleware

import (
	"errors"

	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// OK writes { "data": payload } with 200.
func OK(c *gin.Context, data any) {
	c.JSON(200, gin.H{"data": data})
}

// Created writes { "data": payload } with 201.
func Created(c *gin.Context, data any) {
	c.JSON(201, gin.H{"data": data})
}

// NoContent writes 204.
func NoContent(c *gin.Context) {
	c.Status(204)
}

// Paginated writes { "data": [...], "meta": {...} }.
func Paginated(c *gin.Context, data any, cursor string, hasMore bool, limit int) {
	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{
			"cursor":   cursor,
			"has_more": hasMore,
			"limit":    limit,
		},
	})
}

// Err maps an error to the appropriate HTTP response.
func Err(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	c.AbortWithStatusJSON(500, gin.H{"error": apperror.ErrInternal})
}
