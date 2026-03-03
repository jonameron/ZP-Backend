package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type InviteRepository struct {
	db *gorm.DB
}

func NewInviteRepository(db *gorm.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(invite *models.Invite) error {
	return r.db.Create(invite).Error
}

func (r *InviteRepository) FindByToken(token string) (*models.Invite, *models.Event, *models.User, error) {
	var invite models.Invite
	if err := r.db.Where("token = ?", token).First(&invite).Error; err != nil {
		return nil, nil, nil, err
	}

	var event models.Event
	if err := r.db.First(&event, invite.EventRef).Error; err != nil {
		return nil, nil, nil, err
	}

	var host models.User
	if err := r.db.First(&host, invite.HostRef).Error; err != nil {
		return nil, nil, nil, err
	}

	return &invite, &event, &host, nil
}

func (r *InviteRepository) Update(invite *models.Invite) error {
	return r.db.Save(invite).Error
}

func (r *InviteRepository) CountAcceptedByHost(hostID uint) int64 {
	var count int64
	r.db.Model(&models.Invite{}).Where("host_ref = ? AND status = ?", hostID, "Accepted").Count(&count)
	return count
}
