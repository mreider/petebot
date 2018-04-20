package main

const HELP_MESSAGE = `*HELP* You're looking at it
*CLUB* Get the club id you are watching
*CLUB [strava club id]*  Watch a Strava club
*CLUB RESET* Stop watching a Strava club
*UNIT [mi miles, km kilometers, m meters]* Change club units
*RECENT [# of activities]* Show recent activities
*MESSAGE [text]* Customize your activity message
*MESSAGE RESET* Set activity message back to default
*DONATE* Learn how to donate
_Confused? See pete.fit for docs_

`

type HelpCommand struct {
}

func (cmd *HelpCommand) Name() string {
	return "help"
}

func (cmd *HelpCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	return HELP_MESSAGE, nil
}
