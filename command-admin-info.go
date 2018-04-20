package main

import (
	"fmt"
)

const (
	USERS_TOTAL_AND_STRAVA    = "Users: %d (unique strava users: %d)"
	USERS_TOTAL               = "Users: %d"
	USERS_UNLOCKED            = "Unlocked users: %d"
	TEAMS_TOTAL_AND_MONITORED = "Teams: %d (monitored: %d)"
	TEAMS_TOTAL               = "Teams: %d"
	CLUBS_TOTAL               = "Clubs: %d"
	RATE_LIMITS_UNKNOWN       = "Rate usage/limits are unknown - Strava API hasn't be called yet"
	RATE_LIMITS               = "Rate usage/limits at %s: 15-minute usage - %d/%d requests, day usage - %d/%d requests"
)

type AdminInfoCommand struct {
}

func (cmd *AdminInfoCommand) Name() string {
	return "admin_info"
}

func (cmd *AdminInfoCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	checkResult, err := executor.checkFromAdmin(message)
	if err != nil {
		return "", err
	}
	if checkResult != "" {
		return checkResult, nil
	}

	totalUsersCount, err := executor.repo.AccessDetails.Count()
	if err != nil {
		return "", err
	}

	stravaUsersCount, err := executor.repo.AccessDetails.CountStravaUsers()
	if err != nil {
		return "", err
	}

	unlockedUsersCount, err := executor.repo.UserDetails.CountUnlocked()
	if err != nil {
		return "", err
	}

	totalTeamsCount, err := executor.repo.AccessDetails.CountTeams()
	if err != nil {
		return "", err
	}

	monitoredTeamsCount, err := executor.repo.JobDetails.CountTeams()
	if err != nil {
		return "", err
	}

	clubsCount, err := executor.repo.JobDetails.CountClubs()
	if err != nil {
		return "", err
	}

	var usersCountStr string
	if stravaUsersCount != totalUsersCount {
		usersCountStr = fmt.Sprintf(USERS_TOTAL_AND_STRAVA, totalUsersCount, stravaUsersCount)
	} else {
		usersCountStr = fmt.Sprintf(USERS_TOTAL, totalUsersCount)
	}

	unlockedUsersCountStr := fmt.Sprintf(USERS_UNLOCKED, unlockedUsersCount)

	var teamsCountStr string
	if monitoredTeamsCount != totalTeamsCount {
		teamsCountStr = fmt.Sprintf(TEAMS_TOTAL_AND_MONITORED, totalTeamsCount, monitoredTeamsCount)
	} else {
		teamsCountStr = fmt.Sprintf(TEAMS_TOTAL, totalTeamsCount)
	}

	clubsCountStr := fmt.Sprintf(CLUBS_TOTAL, clubsCount)

	var rateLimitisStr string
	requestTime, limitShort, limitLong, usageShort, usageLong := executor.strava.GetRateLimits()
	if requestTime.IsZero() {
		rateLimitisStr = RATE_LIMITS_UNKNOWN
	} else {
		rateLimitisStr = fmt.Sprintf(RATE_LIMITS,
			requestTime.Format("02-Jan-06 15:04:05"), usageShort, limitShort, usageLong, limitLong)
	}

	return usersCountStr + "\n" + unlockedUsersCountStr + "\n" + teamsCountStr + "\n" + clubsCountStr + "\n\n" + rateLimitisStr, nil
}
