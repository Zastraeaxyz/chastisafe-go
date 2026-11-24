package chastisafe

import "testing"

func TestLogEntryText(t *testing.T) {
	tests := []struct {
		name  string
		entry LogEntry
		want  string
	}{
		{
			name:  "keyholder unlocked",
			entry: LogEntry{Actor: 1, EventType: 12},
			want:  "Keyholder unlocked",
		},
		{
			name:  "lockee picked red",
			entry: LogEntry{Actor: 2, EventType: 6, Card: "r", Amount: 1},
			want:  "Lockee picked 1 red",
		},
		{
			name:  "lockee picked yellow three",
			entry: LogEntry{Actor: 2, EventType: 6, Card: "3", Amount: 1},
			want:  "Lockee picked 1 +3 red",
		},
		{
			name:  "dice pick",
			entry: LogEntry{Actor: 2, EventType: 6, Card: "i", Amount: 4},
			want:  "Lockee picked 1 dice. Dice thrown was 4",
		},
		{
			name:  "loaded",
			entry: LogEntry{Actor: 2, EventType: 39},
			want:  "Lockee loaded the lock",
		},
		{
			name:  "system added greens",
			entry: LogEntry{Actor: 0, EventType: 15, Card: "g", Amount: 2},
			want:  "System added 2 greens to the deck",
		},
		{
			name:  "timer update",
			entry: LogEntry{Actor: 1, EventType: 19, Amount: 3600},
			want:  "Keyholder updated the timer duration to 1 hour",
		},
		{
			name:  "unknown event",
			entry: LogEntry{Actor: 2, EventType: 999},
			want:  "Lockee unknown event",
		},
		{
			name:  "note",
			entry: LogEntry{Actor: 2, EventType: 52, Info: "hello"},
			want:  "Lockee hello",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := LogEntryText(test.entry); got != test.want {
				t.Fatalf("LogEntryText = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDrawsAvailable(t *testing.T) {
	card := &LoadedLock{Type: "card", Picks: 2}
	if got := card.DrawsAvailable(); got != 2 {
		t.Fatalf("DrawsAvailable = %d, want 2", got)
	}
	timer := &LoadedLock{Type: "timer", Picks: 2}
	if got := timer.DrawsAvailable(); got != 0 {
		t.Fatalf("timer DrawsAvailable = %d, want 0", got)
	}
	empty := &LoadedLock{Type: "card", Picks: 0}
	if got := empty.DrawsAvailable(); got != 0 {
		t.Fatalf("empty DrawsAvailable = %d, want 0", got)
	}
}
