package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	CLUB_WATCHING_CLUB        = "Watching club %s in this channel"
	CLUB_RESET_SUCCESSFUL     = "Ok. No longer watching any clubs in this channel."
	CLUB_ID_SHOULD_BE_INTEGER = "Sorry. Club id's should be integers"
	CLUB_IS_ALREADY_SET       = "Sorry, you're already watching a club in this channel."
	CLUB_DOESNT_EXIST         = "Sorry. That club id doesn't exist or you aren't in the club"
	CLUB_SET_SUCCESSFUL       = "Great! I will post updates from this Strava club"
)

type ClubCommand struct {
}

func (cmd *ClubCommand) Name() string {
	return "club"
}

func (cmd *ClubCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	if len(params) > 1 {
		return COMMAND_TOO_MANY_PARAMETERS, nil
	}

	job, err := executor.repo.JobDetails.Get(message.TeamId, message.ChannelId)
	if err != nil {
		return "", err
	}

	if len(params) == 0 {
		if job == nil {
			return COMMAND_NO_CLUBS_WATCHED, nil
		} else {
			return fmt.Sprintf(CLUB_WATCHING_CLUB, job.ClubId), nil
		}
	}

	if strings.ToLower(params[0]) == "reset" {
		if job != nil {
			err = executor.repo.JobDetails.Delete(job)
			if err != nil {
				return "", err
			}
			executor.jobsExecutor.RemoveJob(job)
		}
		return CLUB_RESET_SUCCESSFUL, nil
	}

	if job != nil {
		return CLUB_IS_ALREADY_SET, nil
	}

	clubId := params[0]
	numericClubId, err := strconv.ParseInt(clubId, 10, 64)
	if err != nil {
		return CLUB_ID_SHOULD_BE_INTEGER, nil
	}

	details, err := executor.repo.AccessDetails.GetForTeam(message.TeamId, message.UserId)
	if err != nil {
		return "", err
	} else if details == nil {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}

	activities, err := executor.strava.GetClubActivities(details, numericClubId)
	if err != nil {
		return CLUB_DOESNT_EXIST, nil
	}

	job = &JobDetails{TeamId: message.TeamId, ClubId: clubId, ChannelId: message.ChannelId, SlackUserId: message.UserId}
	err = executor.repo.JobDetails.Create(job)
	if err != nil {
		return "", err
	}

	executor.jobsExecutor.AddJob(job)
	go func() {
		time.Sleep(3 * time.Second)
		executor.poster.PostActivities(clubId, activities, []*JobDetails{job})
	}()
	return CLUB_SET_SUCCESSFUL, nil
}
