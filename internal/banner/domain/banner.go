package domain

import (
	"GAMERS-BE/internal/global/exception"
	"time"
)

const (
	MaxBannerCount = 5
)

type MainBanner struct {
	ID           int64     `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	ImageKey     string    `gorm:"column:image_key;type:varchar(512);not null" json:"image_key"`
	Title        *string   `gorm:"column:title;type:varchar(255)" json:"title,omitempty"`
	LinkURL      *string   `gorm:"column:link_url;type:varchar(512)" json:"link_url,omitempty"`
	DisplayOrder int       `gorm:"column:display_order;default:0" json:"display_order"`
	IsActive     bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime" json:"created_at"`
	ModifiedAt   time.Time `gorm:"column:modified_at;type:timestamp;autoUpdateTime" json:"modified_at"`
}

func (b *MainBanner) TableName() string {
	return "main_banners"
}

func NewMainBanner(imageKey string, title, linkURL *string, displayOrder int, isActive bool) (*MainBanner, error) {
	if imageKey == "" {
		return nil, exception.ErrFileRequired
	}

	return &MainBanner{
		ImageKey:     imageKey,
		Title:        title,
		LinkURL:      linkURL,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
	}, nil
}

func (b *MainBanner) Update(imageKey *string, title, linkURL *string, displayOrder *int, isActive *bool) {
	if imageKey != nil && *imageKey != "" {
		b.ImageKey = *imageKey
	}
	if title != nil {
		b.Title = title
	}
	if linkURL != nil {
		b.LinkURL = linkURL
	}
	if displayOrder != nil {
		b.DisplayOrder = *displayOrder
	}
	if isActive != nil {
		b.IsActive = *isActive
	}
}
