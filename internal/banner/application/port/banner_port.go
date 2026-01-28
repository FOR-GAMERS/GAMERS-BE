package port

import "GAMERS-BE/internal/banner/domain"

type BannerPort interface {
	// FindById retrieves a banner by its ID
	FindById(id int64) (*domain.MainBanner, error)

	// FindAllActive retrieves all active banners ordered by display_order
	FindAllActive() ([]*domain.MainBanner, error)

	// FindAll retrieves all banners ordered by display_order
	FindAll() ([]*domain.MainBanner, error)

	// Count returns the total number of banners
	Count() (int64, error)

	// Save creates a new banner
	Save(banner *domain.MainBanner) error

	// Update updates an existing banner
	Update(banner *domain.MainBanner) error

	// Delete removes a banner by its ID
	Delete(id int64) error
}
