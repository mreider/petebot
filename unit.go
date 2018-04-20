package main

import (
	"math"
	"strings"
)

type Unit struct {
	Name        string
	Aliases     []string
	MeterFactor float64
}

var KnownUnits = []*Unit{
	{Name: "mi", Aliases: []string{"miles", "mile"}, MeterFactor: 1609.344},
	{Name: "km", Aliases: []string{"kilometers", "kilometer"}, MeterFactor: 1000},
	{Name: "ft", Aliases: []string{"feet", "foot"}, MeterFactor: 0.3048},
	{Name: "m", Aliases: []string{"meters", "meter"}, MeterFactor: 1},
}

var DefaultUnit = KnownUnits[0]

func GetKnownUnit(unitStr string) *Unit {
	unitStr = strings.TrimSpace(strings.ToLower(unitStr))
	for _, unit := range KnownUnits {
		if unit.Name == unitStr {
			return unit
		}
		for _, alias := range unit.Aliases {
			if alias == unitStr {
				return unit
			}
		}
	}
	return nil
}

func ConvertToUnit(value float64, unitName string) float64 {
	unit := GetKnownUnit(unitName)
	if unit != nil {
		return math.Floor((value/unit.MeterFactor)*10+0.5) / 10
	} else {
		return value
	}
}

func ConvertToSpeed(value float64, unitName string) float64 {
	unit := GetKnownUnit(unitName)
	if unit != nil {
		if unit.Name == "m" || unit.Name == "ft" {
			return math.Floor((value/unit.MeterFactor)*10+0.5) / 10
		} else {
			return math.Floor((value*60*60)/unit.MeterFactor*10+0.5) / 10
		}
	} else {
		return value
	}
}

func ConvertToPace(value float64, unitName string) float64 {
	unit := GetKnownUnit(unitName)
	if unit != nil {
		if unit.Name == "m" {
			return math.Floor((unit.MeterFactor/value)*100+0.5) / 100
		} else if unit.Name == "ft" {
			return math.Floor((unit.MeterFactor/value)*10+0.5) / 10
		} else {
			return math.Floor(unit.MeterFactor/(value*60)*10+0.5) / 10
		}
	} else {
		return value
	}
}
