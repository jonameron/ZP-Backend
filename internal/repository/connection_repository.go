package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type ConnectionRepository struct {
	db *gorm.DB
}

func NewConnectionRepository(db *gorm.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

func (r *ConnectionRepository) Create(conn *models.Connection) error {
	return r.db.Create(conn).Error
}

func (r *ConnectionRepository) FindByID(id uint) (*models.Connection, error) {
	var conn models.Connection
	err := r.db.First(&conn, id).Error
	return &conn, err
}

func (r *ConnectionRepository) Update(conn *models.Connection) error {
	return r.db.Save(conn).Error
}

func (r *ConnectionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Connection{}, id).Error
}

// FindExisting checks if a connection already exists between two users (in either direction)
func (r *ConnectionRepository) FindExisting(userID, connectedUserID uint) (*models.Connection, error) {
	var conn models.Connection
	err := r.db.Where(
		"(user_ref = ? AND connected_ref = ?) OR (user_ref = ? AND connected_ref = ?)",
		userID, connectedUserID, connectedUserID, userID,
	).First(&conn).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &conn, err
}

// FindAcceptedConnections returns all accepted connections for a user
func (r *ConnectionRepository) FindAcceptedConnections(userID uint) ([]models.Connection, error) {
	var conns []models.Connection
	err := r.db.Where(
		"(user_ref = ? OR connected_ref = ?) AND status = ?",
		userID, userID, "accepted",
	).Find(&conns).Error
	return conns, err
}

// FindPendingForUser returns connection requests sent TO this user
func (r *ConnectionRepository) FindPendingForUser(userID uint) ([]models.Connection, error) {
	var conns []models.Connection
	err := r.db.Where("connected_ref = ? AND status = ?", userID, "pending").Find(&conns).Error
	return conns, err
}

// GetConnectedUserIDs returns all user IDs that are accepted connections of the given user
func (r *ConnectionRepository) GetConnectedUserIDs(userID uint) ([]uint, error) {
	conns, err := r.FindAcceptedConnections(userID)
	if err != nil {
		return nil, err
	}

	var ids []uint
	for _, c := range conns {
		if c.UserRef == userID {
			ids = append(ids, c.ConnectedRef)
		} else {
			ids = append(ids, c.UserRef)
		}
	}
	return ids, nil
}
