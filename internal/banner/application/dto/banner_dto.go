package dto

import (
	"GAMERS-BE/internal/banner/domain"
	"time"
)

// CreateBannerRequest represents the request body for creating a banner
type CreateBannerRequest struct {
	ImageKey     string  `json:"image_key" binding:"required"`
	Title        *string `json:"title"`
	LinkURL      *string `json:"link_url"`
	DisplayOrder int     `json:"display_order" binding:"min=0"`
	IsActive     *bool   `json:"is_active"`
}

// UpdateBannerRequest represents the request body for updating a banner
type UpdateBannerRequest struct {
	ImageKey     *string `json:"image_key"`
	Title        *string `json:"title"`
	LinkURL      *string `json:"link_url"`
	DisplayOrder *int    `json:"display_order" binding:"omitempty,min=0"`
	IsActive     *bool   `json:"is_active"`
}

// BannerResponse represents the response body for a banner
type BannerResponse struct {
	ID           int64     `json:"id"`
	ImageKey     string    `json:"image_key"`
	Title        *string   `json:"title,omitempty"`
	LinkURL      *string   `json:"link_url,omitempty"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

// BannerListResponse represents the response body for a list of banners
type BannerListResponse struct {
	Banners []*BannerResponse `json:"banners"`
	Total   int               `json:"total"`
}

// ToBannerResponse converts a domain.MainBanner to BannerResponse
func ToBannerResponse(banner *domain.MainBanner) *BannerResponse {
	return &BannerResponse{
		ID:           banner.ID,
		ImageKey:     banner.ImageKey,
		Title:        banner.Title,
		LinkURL:      banner.LinkURL,
		DisplayOrder: banner.DisplayOrder,
		IsActive:     banner.IsActive,
		CreatedAt:    banner.CreatedAt,
		ModifiedAt:   banner.ModifiedAt,
	}
}

// ToBannerListResponse converts a slice of domain.MainBanner to BannerListResponse
func ToBannerListResponse(banners []*domain.MainBanner) *BannerListResponse {
	responses := make([]*BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = ToBannerResponse(banner)
	}
	return &BannerListResponse{
		Banners: responses,
		Total:   len(responses),
	}
}
