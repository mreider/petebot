package main

import (
	"github.com/jinzhu/gorm"
)

type PendingUnlocksRepository struct {
	db *gorm.DB
}

func NewPendingUnlocksRepository(db *gorm.DB) *PendingUnlocksRepository {
	return &PendingUnlocksRepository{db: db}
}

func (repo *PendingUnlocksRepository) Get(slackUserId string) (*PendingUnlock, error) {
	pendingUnlock := new(PendingUnlock)
	res := repo.db.First(pendingUnlock, &PendingUnlock{SlackUserId: slackUserId})
	if res.Error == nil {
		return pendingUnlock, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *PendingUnlocksRepository) Create(slackUserId string) error {
	pendingUnlock := &PendingUnlock{SlackUserId: slackUserId}
	res := repo.db.Create(pendingUnlock)
	return res.Error
}

func (repo *PendingUnlocksRepository) Delete(pendingUnlock *PendingUnlock) error {
	res := repo.db.Delete(pendingUnlock)
	return res.Error
}
