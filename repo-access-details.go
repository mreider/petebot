package main

import (
	"github.com/jinzhu/gorm"
)

type AccessDetailsRepository struct {
	db *gorm.DB
}

func NewAccessDetailsRepository(db *gorm.DB) *AccessDetailsRepository {
	return &AccessDetailsRepository{db: db}
}

func (repo *AccessDetailsRepository) GetForTeam(teamId, slackUserId string) (*AccessDetails, error) {
	accessDetails, err := repo.GetForUser(teamId, slackUserId)
	if err != nil {
		return nil, err
	} else if accessDetails != nil {
		return accessDetails, nil
	}

	accessDetails = new(AccessDetails)
	res := repo.db.Last(accessDetails, &AccessDetails{TeamId: teamId}) // Use the most recent credentials
	if res.Error == nil {
		return accessDetails, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *AccessDetailsRepository) GetForUser(teamId, slackUserId string) (*AccessDetails, error) {
	accessDetails := new(AccessDetails)
	res := repo.db.Last(accessDetails, &AccessDetails{TeamId: teamId, SlackUserId: slackUserId}) // Use the most recent credentials
	if res.Error == nil {
		return accessDetails, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *AccessDetailsRepository) GetByStravaUserId(stravaUserId int) (*AccessDetails, error) {
	accessDetails := new(AccessDetails)
	res := repo.db.Last(accessDetails, &AccessDetails{StravaUserId: stravaUserId}) // Use the most recent credentials
	if res.Error == nil {
		return accessDetails, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *AccessDetailsRepository) List() ([]*AccessDetails, error) {
	var accessDetails []*AccessDetails
	res := repo.db.Find(&accessDetails)
	if res.Error != nil {
		return nil, res.Error
	}
	return accessDetails, nil
}

func (repo *AccessDetailsRepository) Create(accessDetails *AccessDetails) error {
	res := repo.db.Create(accessDetails)
	return res.Error
}

func (repo *AccessDetailsRepository) Update(accessDetails *AccessDetails, values map[string]interface{}) error {
	res := repo.db.Model(accessDetails).Updates(values)
	return res.Error
}

func (repo *AccessDetailsRepository) Count() (int, error) {
	var count int
	res := repo.db.Model(&AccessDetails{}).Count(&count)
	if res.Error != nil {
		return -1, res.Error
	}
	return count, nil
}

func (repo *AccessDetailsRepository) CountStravaUsers() (int, error) {
	var result struct{ Counter int }
	res := repo.db.Raw("select count(distinct strava_user_id) counter from access_details;").Scan(&result)
	if res.Error != nil {
		return -1, res.Error
	}
	return result.Counter, nil
}

func (repo *AccessDetailsRepository) CountTeams() (int, error) {
	var result struct{ Counter int }
	res := repo.db.Raw("select count(distinct team_id) counter from access_details;").Scan(&result)
	if res.Error != nil {
		return -1, res.Error
	}
	return result.Counter, nil
}

func (repo *AccessDetailsRepository) Remove(stravaUserId int) error {
	res := repo.db.Exec("delete from access_details where strava_user_id = ?;", stravaUserId)
	return res.Error
}
