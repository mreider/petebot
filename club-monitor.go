package main

import (
	"log"
	"sort"
	"strconv"
	"time"
)

// Monitor that monitors Strava club activities.
type ClubMonitor struct {
	strava               *Strava
	clubId               string
	numericClubId        int64
	initialJob           *JobDetails
	addedJobs            chan *JobDetails
	removedJobs          chan *JobDetails
	commonInterval       time.Duration
	unlockedInterval     time.Duration
	startDelay           time.Duration
	empty                chan<- string
	poster               *ActivitiesPoster
	teamsUnlockInfo      *TeamsUnlockInfo
	lastCommonPostTime   time.Time
	lastUnlockedPostTime time.Time
}

// Create new club monitor.
func NewClubMonitor(strava *Strava, clubId string, initialJob *JobDetails, commonInterval, unlockedInterval, startDelay time.Duration, empty chan<- string, poster *ActivitiesPoster, teamsUnlockInfo *TeamsUnlockInfo) *ClubMonitor {
	numericClubId, _ := strconv.ParseInt(clubId, 10, 64)
	return &ClubMonitor{
		strava:           strava,
		clubId:           clubId,
		numericClubId:    numericClubId,
		initialJob:       initialJob,
		addedJobs:        make(chan *JobDetails, 10),
		removedJobs:      make(chan *JobDetails, 10),
		commonInterval:   commonInterval,
		unlockedInterval: unlockedInterval,
		startDelay:       startDelay,
		empty:            empty,
		poster:           poster,
		teamsUnlockInfo:  teamsUnlockInfo,
	}
}

// Start monitoring a job.
func (m *ClubMonitor) AddJob(job *JobDetails) {
	m.addedJobs <- job
}

// Stop monitoring a job.
func (m *ClubMonitor) RemoveJob(job *JobDetails) {
	m.removedJobs <- job
}

// Run monitoring.
func (m *ClubMonitor) Run() {
	if m.startDelay > 0 {
		time.Sleep(m.startDelay)
	}

	jobs := make([]*JobDetails, 0, 10)
	jobs = append(jobs, m.initialJob)

	if m.startDelay > 0 {
		m.postClubActivities(jobs)
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			m.postClubActivities(jobs)
		case job := <-m.addedJobs:
			jobs = append(jobs, job)
		case job := <-m.removedJobs:
			newJobs := make([]*JobDetails, 0, len(jobs))
			for _, j := range jobs {
				if j.JobId != job.JobId {
					newJobs = append(newJobs, j)
				}
			}
			jobs = newJobs
			if len(jobs) == 0 {
				m.empty <- m.clubId
				break loop
			}
		}
	}
}

// Post club activities to slack channels.
func (m *ClubMonitor) postClubActivities(jobs []*JobDetails) {
	if len(jobs) == 0 {
		return
	}

	commonJobs := make([]*JobDetails, 0, len(jobs))
	unlockedJobs := make([]*JobDetails, 0, len(jobs))
	for _, job := range jobs {
		if m.teamsUnlockInfo.HasUnlockedUser(job.TeamId) {
			unlockedJobs = append(unlockedJobs, job)
		} else {
			commonJobs = append(commonJobs, job)
		}
	}

	currentTime := time.Now()
	var todoJobs = make([]*JobDetails, 0, len(jobs))

	if currentTime.After(m.lastUnlockedPostTime.Add(m.unlockedInterval).Add(-1 * time.Second)) {
		todoJobs = append(todoJobs, unlockedJobs...)
		m.lastUnlockedPostTime = currentTime
	}

	if currentTime.After(m.lastCommonPostTime.Add(m.commonInterval).Add(-1 * time.Second)) {
		todoJobs = append(todoJobs, commonJobs...)
		m.lastCommonPostTime = currentTime
	}

	if len(todoJobs) == 0 {
		return
	}

	sort.Slice(todoJobs, func(i, j int) bool { return todoJobs[i].Id < todoJobs[j].Id })
	lastJob := todoJobs[len(todoJobs)-1] // Use the most recent job to obtain credentials.
	details, err := m.poster.repo.AccessDetails.GetForTeam(lastJob.TeamId, lastJob.SlackUserId)
	if err != nil {
		log.Printf("Can't load access details from the database: %v\n", err)
		return
	} else if details == nil {
		log.Printf("Team %s. %s\n", lastJob.TeamId, COMMAND_STRAVA_NOT_CONNECTED)
		return
	}

	activities, err := m.strava.GetClubActivities(details, m.numericClubId)
	if err != nil {
		log.Printf("Error when try to get club activities: %v\n", err)
	} else {
		m.poster.PostActivities(m.clubId, activities, todoJobs)
	}
}
