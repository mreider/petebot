package main

const (
	UNLOCK_CODE_SHOULD_BE_SPECIFIED = "Unlock code is required"
	UNLOCK_CODE_IS_INVALID          = "Sorry, that code is invalid"
	UNLOCK_CODE_IS_ALREADY_USED     = "Sorry, that code has already been used"
	UNLOCK_SUCCESSFUL               = "Great! Your account is unlocked."
)

type UnlockCommand struct {
}

func (cmd *UnlockCommand) Name() string {
	return "unlock"
}

func (cmd *UnlockCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	if len(params) != 1 {
		return UNLOCK_CODE_SHOULD_BE_SPECIFIED, nil
	}

	unlockCode := params[0]

	details, err := executor.repo.AccessDetails.GetForUser(message.TeamId, message.UserId)
	if err != nil {
		return "", err
	} else if details == nil {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}

	if details.StravaUserId == 0 {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}

	sentCode, err := executor.repo.SentCodes.Get(unlockCode)
	if err != nil {
		return "", err
	} else if sentCode == nil {
		return UNLOCK_CODE_IS_INVALID, nil
	} else if sentCode.Used {
		return UNLOCK_CODE_IS_ALREADY_USED, nil
	} else {
		err := executor.repo.SentCodes.Update(sentCode, map[string]interface{}{"Used": true})
		if err != nil {
			return "", err
		}

		err = executor.unlockUser(details.StravaUserId, unlockCode, message.TeamId)
		if err != nil {
			return "", err
		}
		return UNLOCK_SUCCESSFUL, nil
	}
}
