package main

const (
	COMMAND_NO_CLUBS_WATCHED     = "No clubs watched. Type CLUB [club id] to watch one"
	COMMAND_TOO_MANY_PARAMETERS  = "Too many parameters"
	COMMAND_STRAVA_NOT_CONNECTED = "You must connect to strava & slack at https://pete.fit"
	//COMMAND_LOCKED_ACCOUNT       = "Sorry. You need to unlock your account to use this command"
	COMMAND_SHOULD_BE_ADMIN      = "Sorry. You should be the Administrator"
)

type Command interface {
	Name() string
	Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error)
}

type CommandWithAttachment interface {
	ExecuteWithAttachment(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (response, attachmentName string, attacmentContent []byte, err error)
}
