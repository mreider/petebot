package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDurationText(t *testing.T) {
	assert.Equal(t, "0sec", getDurationText(0))
	assert.Equal(t, "9sec", getDurationText(9))
	assert.Equal(t, "59sec", getDurationText(59))
	assert.Equal(t, "1m00sec", getDurationText(60))
	assert.Equal(t, "59m59sec", getDurationText(3599))
	assert.Equal(t, "1h00m00sec", getDurationText(3600))
	assert.Equal(t, "1h07m25sec", getDurationText(3600+7*60+25))
	assert.Equal(t, "101h07m25sec", getDurationText(3600*101+7*60+25))
}
