package main

import "github.com/jinzhu/gorm"

type Repository struct {
	db *gorm.DB

	AccessDetails   *AccessDetailsRepository
	BotDetails      *BotDetailsRepository
	JobDetails      *JobDetailsRepository
	SentDetails     *SentDetailsRepository
	SentCodes       *SentCodesRepository
	ClubDetails     *ClubDetailsRepository
	UserDetails     *UserDetailsRepository
	UserClubDetails *UserClubDetailsRepository
	PendingUnlocks  *PendingUnlocksRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		AccessDetails:   NewAccessDetailsRepository(db),
		BotDetails:      NewBotDetailsRepository(db),
		JobDetails:      NewJobDetailsRepository(db),
		SentDetails:     NewSentDetailsRepository(db),
		SentCodes:       NewSentCodesRepository(db),
		ClubDetails:     NewClubDetailsRepository(db),
		UserDetails:     NewUserDetailsRepository(db),
		UserClubDetails: NewUserClubDetailsRepository(db),
		PendingUnlocks:  NewPendingUnlocksRepository(db),
	}
}
