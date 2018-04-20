package main

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"strings"
)

type JobDetailsRepository struct {
	db *gorm.DB
}

func NewJobDetailsRepository(db *gorm.DB) *JobDetailsRepository {
	return &JobDetailsRepository{db: db}
}

func (repo *JobDetailsRepository) Get(teamId, channelId string) (*JobDetails, error) {
	job := new(JobDetails)
	res := repo.db.Last(job, &JobDetails{TeamId: teamId, ChannelId: channelId})
	if res.Error == nil {
		return job, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *JobDetailsRepository) ListAll() ([]*JobDetails, error) {
	var jobs []*JobDetails
	res := repo.db.Find(&jobs)
	if res.Error != nil {
		return nil, res.Error
	}
	return jobs, nil
}

func (repo *JobDetailsRepository) Create(job *JobDetails) error {
	jobId := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	job.JobId = jobId
	res := repo.db.Create(job)
	return res.Error
}

func (repo *JobDetailsRepository) Delete(job *JobDetails) error {
	res := repo.db.Delete(job)
	return res.Error
}

func (repo *JobDetailsRepository) Update(job *JobDetails, values map[string]interface{}) error {
	res := repo.db.Model(job).Updates(values)
	return res.Error
}

func (repo *JobDetailsRepository) ListTeamsWithUnlockedUsers() (map[string]bool, error) {
	rows, err := repo.db.Raw(
		`select distinct ad.team_id
		from access_details ad
		where exists(select null from user_details ud where ud.strava_user_id = ad.strava_user_id
			-- and ud.unlock_code is not null and ud.unlock_code != ''
			);`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var teamId string
		err := rows.Scan(&teamId)
		if err != nil {
			return nil, err
		}
		result[teamId] = true
	}
	return result, nil
}

func (repo *JobDetailsRepository) CountTeams() (int, error) {
	var result struct{ Counter int }
	res := repo.db.Raw("select count(distinct team_id) counter from job_details;").Scan(&result)
	if res.Error != nil {
		return -1, res.Error
	}
	return result.Counter, nil
}

func (repo *JobDetailsRepository) CountClubs() (int, error) {
	var result struct{ Counter int }
	res := repo.db.Raw("select count(distinct club_id) counter from job_details;").Scan(&result)
	if res.Error != nil {
		return -1, res.Error
	}
	return result.Counter, nil
}
