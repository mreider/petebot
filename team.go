package main

import "sync"

type TeamsUnlockInfo struct {
	unlockedTeams map[string]bool
	mutex         sync.RWMutex
}

func NewTeamsUnlockInfo(unlockedTeams map[string]bool) *TeamsUnlockInfo {
	return &TeamsUnlockInfo{
		unlockedTeams: unlockedTeams,
	}
}

func (t *TeamsUnlockInfo) HasUnlockedUser(teamId string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.unlockedTeams[teamId]
}

func (t *TeamsUnlockInfo) AddUnlockUser(teamId string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.unlockedTeams[teamId] = true
}
