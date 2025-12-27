package command

import "GAMERS-BE/internal/user/domain"

type ProfileCommandPort interface {
	Save(p *domain.Profile) error
	Update(p *domain.Profile) error
	DeleteById(id int64) error
}
