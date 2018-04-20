package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"strconv"
	"strings"
)

const (
	ADMIN_USERS_UNKNOWN_SUBCOMMAND          = "Unknown subcommand"
	ADMIN_USERS_USER_ID_SHOULD_BE_SPECIFIED = "USER ID should be specified"
	ADMIN_USERS_ID_SHOULD_BE_INTEGER        = "Sorry. Users id's should be integers"
	ADMIN_USERS_UNEXISTING_USER             = "The user doesn't exists or isn't connected to Petebot"
	ADMIN_USERS_USER_REMOVED                = "The user is removed"
)

type AdminUsersCommand struct {
}

func (cmd *AdminUsersCommand) Name() string {
	return "admin_users"
}

func (cmd *AdminUsersCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	response, _, _, err := cmd.ExecuteWithAttachment(params, message, executor)
	return response, err
}

func (cmd *AdminUsersCommand) ExecuteWithAttachment(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (response, attachmentName string, attachmentContent []byte, err error) {
	checkResult, err := executor.checkFromAdmin(message)
	if err != nil {
		return "", "", nil, err
	}
	if checkResult != "" {
		return checkResult, "", nil, nil
	}

	if len(params) == 0 {
		return cmd.listUsers(message, executor)
	}

	if strings.ToLower(params[0]) == "remove" {
		response, err := cmd.removeUsers(params[1:], message, executor)
		return response, "", nil, err
	}

	return ADMIN_USERS_UNKNOWN_SUBCOMMAND, "", nil, nil
}

func (cmd *AdminUsersCommand) listUsers(message *IncomingSlackMessage, executor *CommandExecutor) (response, attachmentName string, attachmentContent []byte, err error) {
	userDetails, err := executor.repo.UserDetails.List()
	if err != nil {
		return "", "", nil, err
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	csvWriter := csv.NewWriter(writer)
	csvWriter.Write([]string{"UserId", "UserName", "Status"})
	for _, user := range userDetails {
		status := ""
		if user.UnlockCode != "" {
			status = "Unlocked"
		}

		csvWriter.Write([]string{strconv.Itoa(user.StravaUserId), user.UserName, status})
	}
	csvWriter.Flush()
	writer.Flush()

	return "", "users.csv", buf.Bytes(), nil
}

func (cmd *AdminUsersCommand) removeUsers(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	checkResult, err := executor.checkFromAdmin(message)
	if err != nil {
		return "", err
	}
	if checkResult != "" {
		return checkResult, nil
	}

	if len(params) == 0 {
		return ADMIN_USERS_USER_ID_SHOULD_BE_SPECIFIED, nil
	}

	if len(params) > 1 {
		return COMMAND_TOO_MANY_PARAMETERS, nil
	}

	userId, err := strconv.Atoi(params[0])
	if err != nil {
		return ADMIN_USERS_ID_SHOULD_BE_INTEGER, nil
	}

	userDetails, err := executor.repo.UserDetails.Get(userId)
	if err != nil {
		return "", err
	}
	if userDetails == nil {
		return ADMIN_USERS_UNEXISTING_USER, nil
	}

	err = executor.repo.UserClubDetails.Remove(userId)
	if err != nil {
		return "", err
	}

	err = executor.repo.AccessDetails.Remove(userId)
	if err != nil {
		return "", err
	}

	err = executor.repo.UserDetails.Delete(userDetails)
	if err != nil {
		return "", err
	}

	return ADMIN_USERS_USER_REMOVED, nil
}
