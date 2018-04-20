package main

import (
	"sort"
	"strconv"
)

const (
	DEFAULT_RECENT_COUNT            = 10
	MAX_RECENT_COUNT                = 200
	RECENT_NUMBER_SHOULD_BE_INTEGER = "Sorry. Number of activities should be integer"
	RECENT_NO_ACTIVITIES            = "No recent activities"
)

type RecentCommand struct {
}

func (cmd *RecentCommand) Name() string {
	return "recent"
}

func (cmd *RecentCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	if len(params) > 1 {
		return COMMAND_TOO_MANY_PARAMETERS, nil
	}

	recentCount := DEFAULT_RECENT_COUNT
	var err error
	if len(params) == 1 {
		recentCount, err = strconv.Atoi(params[0])
		if err != nil {
			return RECENT_NUMBER_SHOULD_BE_INTEGER, nil
		} else if recentCount <= 0 {
			return "", nil
		}
	}

	if recentCount > MAX_RECENT_COUNT {
		recentCount = MAX_RECENT_COUNT
	}

	jobDetails, err := executor.repo.JobDetails.Get(message.TeamId, message.ChannelId)
	if err != nil {
		return "", err
	}

	if jobDetails == nil {
		return COMMAND_NO_CLUBS_WATCHED, nil
	}

	clubId := jobDetails.ClubId
	numericClubId, _ := strconv.ParseInt(clubId, 10, 64)

	details, err := executor.repo.AccessDetails.GetForTeam(message.TeamId, message.UserId)
	if err != nil {
		return "", err
	} else if details == nil {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}

	activities, err := executor.strava.GetClubActivities(details, numericClubId)
	if err != nil {
		return "", err
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].Id > activities[j].Id })
	count := recentCount
	if len(activities) < count {
		count = len(activities)
	}
	activities = activities[:count]
	sort.Slice(activities, func(i, j int) bool { return activities[i].Id < activities[j].Id })

	if len(activities) == 0 {
		return RECENT_NO_ACTIVITIES, nil
	}

	return executor.activitiesPoster.GetActivitiesText(clubId, activities)
}
