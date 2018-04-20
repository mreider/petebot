package main

import (
	"fmt"
	"strings"
)

const (
	MESSAGE_TEXT_IS_DEFAULT    = "You have default message text"
	MESSAGE_TEXT_IS_CUSTOM     = "Your message text: %s"
	MESSAGE_SUCCESSFUL         = "Message text is set.\nSample response: %s"
	MESSAGE_TEXT_IS_RESET      = "Message text is reset to default"
	MESSAGE_INCORRECT_TEMPLATE = "Incorrect message text: %v"
)

type MessageCommand struct {
}

func (cmd *MessageCommand) Name() string {
	return "message"
}

func (cmd *MessageCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	details, err := executor.repo.AccessDetails.GetForUser(message.TeamId, message.UserId)
	if err != nil {
		return "", err
	} else if details == nil {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}

	if details.StravaUserId == 0 {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}

	//userDetails, err := executor.repo.UserDetails.Get(details.StravaUserId)
	//if err != nil {
	//	return "", err
	//}
	//
	//if userDetails == nil || userDetails.UnlockCode == "" {
	//	return COMMAND_LOCKED_ACCOUNT, nil
	//}

	jobDetails, err := executor.repo.JobDetails.Get(message.TeamId, message.ChannelId)
	if err != nil {
		return "", err
	}

	if jobDetails == nil {
		return COMMAND_NO_CLUBS_WATCHED, nil
	}

	userClubDetails, err := executor.repo.UserClubDetails.Get(details.StravaUserId, jobDetails.ClubId)
	if err != nil {
		return "", err
	}

	if len(params) == 0 {
		if userClubDetails == nil || userClubDetails.MessageText == "" {
			return MESSAGE_TEXT_IS_DEFAULT, nil
		} else {
			return fmt.Sprintf(MESSAGE_TEXT_IS_CUSTOM, userClubDetails.MessageText), nil
		}
	}

	messageText := strings.Join(params, " ")
	if strings.ToLower(messageText) == "reset" {
		if userClubDetails != nil {
			err = executor.repo.UserClubDetails.Update(userClubDetails, map[string]interface{}{"MessageText": ""})
			if err != nil {
				return "", err
			}
		}
		return MESSAGE_TEXT_IS_RESET, nil
	}

	tmpl, err := executor.activityTemplateEngine.Compile(messageText)
	if err != nil {
		return fmt.Sprintf(MESSAGE_INCORRECT_TEMPLATE, err), nil
	}

	if userClubDetails == nil {
		userClubDetails := &UserClubDetails{StravaUserId: details.StravaUserId, ClubId: jobDetails.ClubId, MessageText: messageText}
		err = executor.repo.UserClubDetails.Create(userClubDetails)
		if err != nil {
			return "", err
		}
	} else {
		err = executor.repo.UserClubDetails.Update(userClubDetails, map[string]interface{}{"MessageText": messageText})
		if err != nil {
			return "", err
		}
	}

	sample := executor.activityTemplateEngine.SampleText(tmpl)
	return fmt.Sprintf(MESSAGE_SUCCESSFUL, sample), nil
}
