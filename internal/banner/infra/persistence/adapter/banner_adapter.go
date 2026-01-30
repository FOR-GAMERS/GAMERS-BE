package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/banner/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"errors"

	"gorm.io/gorm"
)

type MySQLBannerAdapter struct {
	db *gorm.DB
}

func NewMySQLBannerAdapter(db *gorm.DB) *MySQLBannerAdapter {
	return &MySQLBannerAdapter{db: db}
}

func (a *MySQLBannerAdapter) FindById(id int64) (*domain.MainBanner, error) {
	var banner domain.MainBanner
	result := a.db.First(&banner, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrBannerNotFound
		}
		return nil, result.Error
	}

	return &banner, nil
}

func (a *MySQLBannerAdapter) FindAllActive() ([]*domain.MainBanner, error) {
	var banners []*domain.MainBanner
	result := a.db.Where("is_active = ?", true).Order("display_order ASC").Find(&banners)

	if result.Error != nil {
		return nil, result.Error
	}

	return banners, nil
}

func (a *MySQLBannerAdapter) FindAll() ([]*domain.MainBanner, error) {
	var banners []*domain.MainBanner
	result := a.db.Order("display_order ASC").Find(&banners)

	if result.Error != nil {
		return nil, result.Error
	}

	return banners, nil
}

func (a *MySQLBannerAdapter) Count() (int64, error) {
	var count int64
	result := a.db.Model(&domain.MainBanner{}).Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

func (a *MySQLBannerAdapter) Save(banner *domain.MainBanner) error {
	return a.db.Create(banner).Error
}

func (a *MySQLBannerAdapter) Update(banner *domain.MainBanner) error {
	return a.db.Save(banner).Error
}

func (a *MySQLBannerAdapter) Delete(id int64) error {
	return a.db.Delete(&domain.MainBanner{}, id).Error
}
