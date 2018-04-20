package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildPlain(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("")
	assert.Nil(t, err)
	assert.Equal(t, "", template)

	template, err = engine.Build("1234 abcd")
	assert.Nil(t, err)
	assert.Equal(t, "1234 abcd", template)
}

func TestBuildFirstLevel(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("it is a {name}. cool. {distance} for {moving_time}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{.Distance}} for {{toTime .MovingTime}}!!!", template)
}

func TestBuildSecondLevel(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("Cool {athlete.firstname}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "Cool {{.Athlete.FirstName}}!!!", template)
}

func TestSampleActivity(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	tmpl, err := engine.Compile("{athlete.firstname} {athlete.lastname} - {distance} for {moving_time}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "Patrick Jane - 4475.4 for 21m43sec!!!", engine.SampleText(tmpl))
}

func TestUnit(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("it is a {name}. cool. {distance_km}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{toUnit .Distance \"km\"}}!!!", template)
}

func TestSampleActivityUnit(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	tmpl, err := engine.Compile("{athlete.firstname} {athlete.lastname} - {distance_km} for {moving_time}. Miles - {distance_mi}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "Patrick Jane - 4.5 for 21m43sec. Miles - 2.8!!!", engine.SampleText(tmpl))
}

func TestUnitAlias(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("it is a {name}. cool. {distance_feet}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{toUnit .Distance \"ft\"}}!!!", template)
}

func TestSampleActivityUnitAlias(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	tmpl, err := engine.Compile("{athlete.firstname} {athlete.lastname} - {distance_kilometer} for {moving_time}. Miles - {distance_miles}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "Patrick Jane - 4.5 for 21m43sec. Miles - 2.8!!!", engine.SampleText(tmpl))
}

func TestBuildFirstLevelError(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("it is a {name}. cool. {moving_timer}!!!")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid field {moving_timer}", template)
}

func TestBuildSecondLevelError(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("Cool {athlete.firstnime}!!!")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid field {athlete.firstnime}", template)
}

func TestSpeedBuild(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("it is a {name}. cool. {average_speed}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{toSpeed .AverageSpeed \"m\"}}!!!", template)

	template, err = engine.Build("it is a {name}. cool. {average_speed_feet}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{toSpeed .AverageSpeed \"ft\"}}!!!", template)
}

func TestSpeedCompile(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	tmpl, err := engine.Compile("it is a {name}. cool. {average_speed}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a Evening Ride. cool. 3.4!!!", engine.SampleText(tmpl))

	tmpl, err = engine.Compile("it is a {name}. cool. {average_speed_feet}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a Evening Ride. cool. 11.2!!!", engine.SampleText(tmpl))

	tmpl, err = engine.Compile("it is a {name}. cool. {average_speed_km}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a Evening Ride. cool. 12.2!!!", engine.SampleText(tmpl))
}

func TestPaceBuild(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	template, err := engine.Build("it is a {name}. cool. {average_pace}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{toPace .AverageSpeed \"m\"}}!!!", template)

	template, err = engine.Build("it is a {name}. cool. {average_pace_feet}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a {{.Name}}. cool. {{toPace .AverageSpeed \"ft\"}}!!!", template)
}

func TestPaceCompile(t *testing.T) {
	engine := NewActivityTemplateEngine(".")

	tmpl, err := engine.Compile("it is a {name}. cool. {average_pace}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a Evening Ride. cool. 0.29!!!", engine.SampleText(tmpl))

	tmpl, err = engine.Compile("it is a {name}. cool. {average_pace_feet}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a Evening Ride. cool. 0.1!!!", engine.SampleText(tmpl))

	tmpl, err = engine.Compile("it is a {name}. cool. {average_pace_km}!!!")
	assert.Nil(t, err)
	assert.Equal(t, "it is a Evening Ride. cool. 4.9!!!", engine.SampleText(tmpl))
}
