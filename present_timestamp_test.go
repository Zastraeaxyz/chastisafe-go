package chastisafe

import (
	"strings"
	"testing"
)

func TestLockDetailTextUsesTimestamps(t *testing.T) {
	lock := LoadedLock{
		ID:                  "lock",
		LockName:            "Test card",
		Type:                "card",
		Keyholder:           "key#0",
		KeyholderID:         "1",
		Lockee:              "lockee#0",
		LockeeID:            "2",
		Picks:               0,
		MaxPicks:            1,
		Wait:                300,
		CheckInPeriod:       1440,
		TimeToNextCheckIn:   3600,
		TimeToHygieneUnlock: 7200,
		CardCount:           17,
	}
	text := LockDetailText(&LockView{Lock: lock, Role: RoleLockee})

	for _, want := range []string{
		"<t:",
		":R>",
		"17 cards",
		"Check-in: every 1 day",
		"next <t:",
		"Next draw: <t:",
		"Hygiene unlock: <t:",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("LockDetailText missing %q:\n%s", want, text)
		}
	}
	for _, bad := range []string{"5 minutes", "8h", "next in "} {
		if strings.Contains(text, bad) {
			t.Errorf("LockDetailText still contains %q:\n%s", bad, text)
		}
	}
}

func TestLockDetailTimerUsesTimestamp(t *testing.T) {
	lock := LoadedLock{
		ID:             "timer",
		LockName:       "Test timer",
		Type:           "timer",
		TimerRemaining: 3600,
	}
	text := LockDetailText(&LockView{Lock: lock, Role: RoleLockee})
	if !strings.Contains(text, "Unlocks: <t:") {
		t.Errorf("timer detail missing unlock timestamp:\n%s", text)
	}
	if !strings.Contains(text, "unlocks <t:") {
		t.Errorf("timer state missing unlock timestamp:\n%s", text)
	}
}

func TestLockStateTextUsesTimestamps(t *testing.T) {
	lock := LoadedLock{
		ID:                  "lock",
		Type:                "card",
		TimeToNextCheckIn:   3600,
		TimeToHygieneUnlock: 7200,
		CardCount:           4,
	}
	text := LockStateText(&lock)
	for _, want := range []string{"check-in <t:", "hygiene <t:", ":R>"} {
		if !strings.Contains(text, want) {
			t.Errorf("LockStateText missing %q:\n%s", want, text)
		}
	}
}
