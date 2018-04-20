package main

import (
	"strconv"
	"strings"
)

const (
	ADMIN_UNLOCK_USER_BE_SPECIFIED             = "USER should be mentioned"
	ADMIN_UNLOCK_SUCCESSFUL                    = "User is unlocked."
	ADMIN_UNLOCK_USER_SHOULD_CONNECT_TO_STRAVA = "User should connect to Strava to finish unlocking."
)

type AdminUnlockCommand struct {
}

func (cmd *AdminUnlockCommand) Name() string {
	return "admin_unlock"
}

func (cmd *AdminUnlockCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	checkResult, err := executor.checkFromAdmin(message)
	if err != nil {
		return "", err
	}
	if checkResult != "" {
		return checkResult, nil
	}

	if len(params) != 1 {
		return ADMIN_UNLOCK_USER_BE_SPECIFIED, nil
	}

	var accessDetails *AccessDetails
	var slackOrStravaUserId string
	if strings.HasPrefix(params[0], "<@") && strings.HasSuffix(params[0], ">") {
		unlockSlackUserId := strings.TrimSuffix(strings.TrimPrefix(params[0], "<@"), ">")
		slackOrStravaUserId = unlockSlackUserId
		accessDetails, err = executor.repo.AccessDetails.GetForUser(message.TeamId, unlockSlackUserId)
	} else {
		stravaUserId, err := strconv.Atoi(params[0])
		if err != nil {
			return ADMIN_UNLOCK_USER_BE_SPECIFIED, nil
		}
		slackOrStravaUserId = params[0]
		accessDetails, err = executor.repo.AccessDetails.GetByStravaUserId(stravaUserId)
	}

	if err != nil {
		return "", err
	}

	if accessDetails != nil {
		err := executor.unlockUser(accessDetails.StravaUserId, "ADMIN", message.TeamId)
		if err != nil {
			return "", err
		}
		return ADMIN_UNLOCK_SUCCESSFUL, nil
	} else {
		pendingUnlock, err := executor.repo.PendingUnlocks.Get(slackOrStravaUserId)
		if err != nil {
			return "", err
		}

		if pendingUnlock == nil {
			err := executor.repo.PendingUnlocks.Create(slackOrStravaUserId)
			if err != nil {
				return "", err
			}
		}
		return ADMIN_UNLOCK_USER_SHOULD_CONNECT_TO_STRAVA, nil
	}
}
