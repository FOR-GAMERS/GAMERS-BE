package application

import (
	"GAMERS-BE/internal/banner/application/dto"
	"GAMERS-BE/internal/banner/application/port"
	"GAMERS-BE/internal/banner/domain"
	"GAMERS-BE/internal/global/exception"
)

type BannerService struct {
	bannerPort port.BannerPort
}

func NewBannerService(bannerPort port.BannerPort) *BannerService {
	return &BannerService{
		bannerPort: bannerPort,
	}
}

// GetActiveBanners retrieves all active banners ordered by display_order
func (s *BannerService) GetActiveBanners() (*dto.BannerListResponse, error) {
	banners, err := s.bannerPort.FindAllActive()
	if err != nil {
		return nil, err
	}
	return dto.ToBannerListResponse(banners), nil
}

// GetAllBanners retrieves all banners (for admin)
func (s *BannerService) GetAllBanners() (*dto.BannerListResponse, error) {
	banners, err := s.bannerPort.FindAll()
	if err != nil {
		return nil, err
	}
	return dto.ToBannerListResponse(banners), nil
}

// GetBannerById retrieves a banner by ID
func (s *BannerService) GetBannerById(id int64) (*dto.BannerResponse, error) {
	banner, err := s.bannerPort.FindById(id)
	if err != nil {
		return nil, err
	}
	if banner == nil {
		return nil, exception.ErrBannerNotFound
	}
	return dto.ToBannerResponse(banner), nil
}

// CreateBanner creates a new banner
func (s *BannerService) CreateBanner(req *dto.CreateBannerRequest) (*dto.BannerResponse, error) {
	// Check banner limit
	count, err := s.bannerPort.Count()
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxBannerCount {
		return nil, exception.ErrBannerLimitExceeded
	}

	// Default isActive to true if not specified
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	banner, err := domain.NewMainBanner(req.ImageKey, req.Title, req.LinkURL, req.DisplayOrder, isActive)
	if err != nil {
		return nil, err
	}

	if err := s.bannerPort.Save(banner); err != nil {
		return nil, err
	}

	return dto.ToBannerResponse(banner), nil
}

// UpdateBanner updates an existing banner
func (s *BannerService) UpdateBanner(id int64, req *dto.UpdateBannerRequest) (*dto.BannerResponse, error) {
	banner, err := s.bannerPort.FindById(id)
	if err != nil {
		return nil, err
	}
	if banner == nil {
		return nil, exception.ErrBannerNotFound
	}

	banner.Update(req.ImageKey, req.Title, req.LinkURL, req.DisplayOrder, req.IsActive)

	if err := s.bannerPort.Update(banner); err != nil {
		return nil, err
	}

	return dto.ToBannerResponse(banner), nil
}

// DeleteBanner deletes a banner by ID
func (s *BannerService) DeleteBanner(id int64) error {
	banner, err := s.bannerPort.FindById(id)
	if err != nil {
		return err
	}
	if banner == nil {
		return exception.ErrBannerNotFound
	}

	return s.bannerPort.Delete(id)
}
