package main

import (
	"log"
	"time"
)

const (
	MIN_START_DELAY_IN_SECONDS = 1
)

// Jobs executor.
type JobsExecutor struct {
	strava           *Strava
	commonInterval   time.Duration
	unlockedInterval time.Duration
	jobs             []*JobDetails
	activitiesPoster *ActivitiesPoster
	teamsUnlockInfo  *TeamsUnlockInfo

	addedJobs   chan *JobDetails
	removedJobs chan *JobDetails

	repo *Repository
}

// Create new job executor.
func NewJobsExecutor(strava *Strava, commonInterval, unlockedInterval time.Duration, activitiesPoster *ActivitiesPoster, teamsUnlockInfo *TeamsUnlockInfo, repo *Repository) *JobsExecutor {
	return &JobsExecutor{
		strava:           strava,
		commonInterval:   commonInterval,
		unlockedInterval: unlockedInterval,
		activitiesPoster: activitiesPoster,
		teamsUnlockInfo:  teamsUnlockInfo,
		repo:             repo,
		addedJobs:        make(chan *JobDetails, 10),
		removedJobs:      make(chan *JobDetails, 10),
	}
}

// Add new job.
func (e *JobsExecutor) AddJob(job *JobDetails) {
	e.addedJobs <- job
}

// Remove a job.
func (e *JobsExecutor) RemoveJob(job *JobDetails) {
	e.removedJobs <- job
}

// Run jobs executor.
func (e *JobsExecutor) Run() {
	clubMonitors := make(map[string]*ClubMonitor)
	emptyMonitors := make(chan string, 10)

	for i, job := range e.loadExistingJobs() {
		startDelay := MIN_START_DELAY_IN_SECONDS + i/5
		e.startJobMonitoring(clubMonitors, job, time.Duration(startDelay)*time.Second, emptyMonitors)
	}

	for {
		select {
		case job := <-e.addedJobs:
			e.startJobMonitoring(clubMonitors, job, 0, emptyMonitors)
		case job := <-e.removedJobs:
			if listener, ok := clubMonitors[job.ClubId]; ok {
				listener.RemoveJob(job)
			}
		case clubId := <-emptyMonitors:
			delete(clubMonitors, clubId)
		}
	}
}

// Start job monitoring.
func (e *JobsExecutor) startJobMonitoring(clubMonitors map[string]*ClubMonitor, job *JobDetails, startDelay time.Duration, emptyMonitors chan<- string) {
	var clubMonitor *ClubMonitor
	var ok bool
	if clubMonitor, ok = clubMonitors[job.ClubId]; !ok {
		clubMonitor = NewClubMonitor(e.strava, job.ClubId, job, e.commonInterval, e.unlockedInterval, startDelay, emptyMonitors, e.activitiesPoster, e.teamsUnlockInfo)
		clubMonitors[job.ClubId] = clubMonitor
		go clubMonitor.Run()
	} else {
		clubMonitor.AddJob(job)
	}
}

// Load existing job details from the database.
func (e *JobsExecutor) loadExistingJobs() []*JobDetails {
	jobs, err := e.repo.JobDetails.ListAll()
	if err != nil {
		log.Printf("Can't load job details from the database: %v\n", err)
	}
	return jobs
}
