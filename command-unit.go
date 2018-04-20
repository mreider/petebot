package main

import (
	"fmt"
)

const (
	UNIT_WATCHING_CLUB_NOT_EXIST = "You are not watching any club. Use CLUB to watch a club"
	UNIT_CLUB_UNIT               = "Unit for the club %s: %s"
	UNIT_UNKNOWN_UNIT            = "Unknown unit"
	UNIT_CLUB_UNIT_IS_SET        = "UNIT is set to %s"
)

type UnitCommand struct {
}

func (cmd *UnitCommand) Name() string {
	return "unit"
}

func (cmd *UnitCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	if len(params) > 2 {
		return COMMAND_TOO_MANY_PARAMETERS, nil
	}

	jobDetails, err := executor.repo.JobDetails.Get(message.TeamId, message.ChannelId)
	if err != nil {
		return "", err
	}

	if jobDetails == nil {
		return UNIT_WATCHING_CLUB_NOT_EXIST, nil
	}

	clubId := jobDetails.ClubId
	if len(params) == 0 {
		clubs, err := executor.repo.ClubDetails.List()
		if err != nil {
			return "", err
		}

		clubUnits := make(map[string]string)
		for _, club := range clubs {
			clubUnits[club.ClubId] = club.Unit
		}

		var clubUnit string
		var ok bool
		if clubUnit, ok = clubUnits[clubId]; !ok {
			clubUnit = DefaultUnit.Name
		}
		return fmt.Sprintf(UNIT_CLUB_UNIT, clubId, clubUnit), nil
	}

	unit := GetKnownUnit(params[0])
	if unit == nil {
		return UNIT_UNKNOWN_UNIT, nil
	}

	clubDetails, err := executor.repo.ClubDetails.Get(clubId)
	if err != nil {
		return "", err
	}

	if clubDetails == nil {
		clubDetails = &ClubDetails{ClubId: clubId, Unit: unit.Name}
		err := executor.repo.ClubDetails.Create(clubDetails)
		if err != nil {
			return "", fmt.Errorf("Can't create club details in the database: %v\n", err)
		}
	} else if clubDetails.Unit != unit.Name {
		err := executor.repo.ClubDetails.Update(clubDetails, map[string]interface{}{"Unit": unit.Name})
		if err != nil {
			return "", fmt.Errorf("Can't save club details to the database: %v\n", err)
		}
	}

	return fmt.Sprintf(UNIT_CLUB_UNIT_IS_SET, unit.Name), nil
}
