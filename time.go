package main

import "fmt"

func getDurationText(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%dsec", seconds)
	} else if seconds < 60*60 {
		minutes := seconds / 60
		seconds = seconds % 60
		return fmt.Sprintf("%dm%02dsec", minutes, seconds)
	} else {
		hours := seconds / (60 * 60)
		minutes := (seconds / 60) % 60
		seconds = seconds % 60
		return fmt.Sprintf("%dh%02dm%02dsec", hours, minutes, seconds)
	}
}
