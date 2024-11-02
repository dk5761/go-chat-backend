package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
	TotalItems  int `json:"totalItems"`
	Limit       int `json:"limit"`
	Offset      int `json:"offset"`
}

// Paginate calculates and returns pagination metadata along with limit and offset values
func Paginate(c *gin.Context, totalItems int) (Pagination, int, int, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		return Pagination{}, 0, 0, err
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		return Pagination{}, 0, 0, err
	}

	currentPage := int(math.Floor(float64(offset)/float64(limit)) + 1)
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	pagination := Pagination{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		Limit:       limit,
		Offset:      offset,
	}

	return pagination, limit, offset, nil
}
