package dto

import (
	"fmt"
	"math"
	"strings"
)

// PaginationRequest represents pagination query parameters
type PaginationRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// SortRequest represents sorting parameters
type SortRequest struct {
	SortBy string `form:"sort_by"`
	Order  string `form:"order"`
}

// NewSortRequest creates a new SortRequest with validation
func NewSortRequest(sortBy, order string, allowedFields []string) *SortRequest {
	// Default values
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	// Validate order
	order = strings.ToLower(order)
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Validate sortBy against allowed fields
	sortBy = strings.ToLower(sortBy)
	isAllowed := false
	for _, field := range allowedFields {
		if sortBy == field {
			isAllowed = true
			break
		}
	}
	if !isAllowed && len(allowedFields) > 0 {
		sortBy = allowedFields[0] // Use first allowed field as default
	}

	return &SortRequest{
		SortBy: sortBy,
		Order:  order,
	}
}

// GetOrderClause returns the SQL ORDER BY clause
func (s *SortRequest) GetOrderClause() string {
	return fmt.Sprintf("%s %s", s.SortBy, strings.ToUpper(s.Order))
}

// NewPaginationRequest creates a new PaginationRequest with default values
func NewPaginationRequest(page, pageSize int) *PaginationRequest {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &PaginationRequest{
		Page:     page,
		PageSize: pageSize,
	}
}

// GetOffset calculates the offset for config queries
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size (limit)
func (p *PaginationRequest) GetLimit() int {
	return p.PageSize
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginationResponse creates a new PaginationResponse
func NewPaginationResponse(data interface{}, page, pageSize int, totalCount int64) *PaginationResponse {
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return &PaginationResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}
