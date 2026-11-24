package chastisafe

import (
	"fmt"
	"strings"
	"time"
)

// LogEntryText renders one lock event as a sentence. The labels and the card
// letters mirror the ChastiSafe web client.
func LogEntryText(entry LogEntry) string {
	var b strings.Builder

	switch entry.EventType {
	case 36:
		b.WriteString("Lock started new task roller")
	case 47:
		b.WriteString("Pillory tally is")
	case 52:
		b.WriteString(logActor(entry))
		if info := strings.TrimSpace(entry.Info); info != "" {
			b.WriteString(" ")
			b.WriteString(info)
		}
	default:
		b.WriteString(logActor(entry))
		if verb := logEventVerb(entry.EventType); verb != "" {
			b.WriteString(" ")
			b.WriteString(verb)
		} else {
			b.WriteString(" unknown event")
		}
	}

	card := logCardName(entry.Card)
	switch entry.EventType {
	case 6:
		if card != "" {
			fmt.Fprintf(&b, " 1 %s", card)
		} else {
			b.WriteString(" a card")
		}
		if entry.Card == "i" && entry.Amount > 0 {
			fmt.Fprintf(&b, ". Dice thrown was %d", entry.Amount)
		}
		if entry.Info != "" {
			fmt.Fprintf(&b, " (%s)", entry.Info)
		}
	case 15, 16, 17, 18:
		if card != "" {
			count := entry.Amount
			if count <= 0 {
				count = 1
			}
			fmt.Fprintf(&b, " %d %s", count, card)
			if count > 1 {
				b.WriteString("s")
			}
			b.WriteString(" to the deck")
		} else {
			b.WriteString(" the deck")
		}
	case 19, 38:
		fmt.Fprintf(&b, " duration to %s", FormatDuration(time.Duration(entry.Amount)*time.Second))
	case 48, 49:
		fmt.Fprintf(&b, " %s", FormatDuration(time.Duration(entry.Amount)*time.Second))
	case 21:
		if entry.Info != "" {
			fmt.Fprintf(&b, ". Code is %s", entry.Info)
		}
	case 28, 29, 30, 31, 32, 33, 36, 41, 42, 43, 45, 51:
		if entry.Info != "" {
			fmt.Fprintf(&b, "; %s", entry.Info)
		}
	case 50:
		if entry.Info != "" {
			fmt.Fprintf(&b, " %s", entry.Info)
		}
	}
	return strings.TrimSpace(b.String())
}

func logActor(entry LogEntry) string {
	switch entry.Actor {
	case 1:
		return "Keyholder"
	case 2:
		return "Lockee"
	}
	if name := strings.TrimSpace(entry.Bot); name != "" {
		return name
	}
	return "System"
}

func logEventVerb(eventType int) string {
	switch eventType {
	case 0:
		return "checkin"
	case 1:
		return "missed checkin"
	case 2:
		return "froze the lock"
	case 3:
		return "unfroze the lock"
	case 4:
		return "hid the lock"
	case 5:
		return "unhid the lock"
	case 6:
		return "picked"
	case 7:
		return "automatic reset"
	case 8:
		return "unlocked the lock"
	case 9, 13:
		return "reset the lock"
	case 10:
		return "did random reset"
	case 11:
		return "did random unlock"
	case 12:
		return "unlocked"
	case 14:
		return "abandoned the lock"
	case 15, 17:
		return "added"
	case 16, 18:
		return "removed"
	case 19, 38:
		return "updated the timer"
	case 20:
		return "asked keyholder to decide"
	case 21:
		return "hygiene unlocked"
	case 22:
		return "hygiene locked"
	case 23:
		return "hygiene locked too late"
	case 24:
		return "changed to cumulative picks"
	case 25:
		return "changed to non-cumulative picks"
	case 26:
		return "changed checkin requirements"
	case 27:
		return "changed hygiene unlock parameters"
	case 28, 29:
		return "updated deck"
	case 30:
		return "updated cards"
	case 31:
		return "updated timer"
	case 32:
		return "started task"
	case 33:
		return "completed task"
	case 34:
		return "updated autoresets"
	case 35:
		return "did not complete task on time"
	case 37:
		return "skipped hygiene unlock"
	case 39:
		return "loaded the lock"
	case 40:
		return "split the lock"
	case 41:
		return "pillory started"
	case 42:
		return "completed pillory task"
	case 43:
		return "updated lock"
	case 44:
		return "updated icon"
	case 45:
		return "assigned task"
	case 46:
		return "trusted keyholder"
	case 48:
		return "added time"
	case 49:
		return "removed time"
	case 50:
		return "sent:"
	case 51:
		return "abandoned task"
	default:
		return ""
	}
}

func logCardName(letter string) string {
	letter = strings.TrimSpace(letter)
	if letter == "" {
		return ""
	}
	if card, ok := CardForLetter(letter[0]); ok {
		switch card {
		case CardYellowP1:
			return "yellow+1"
		case CardYellowP2:
			return "yellow+2"
		case CardYellowP3:
			return "yellow+3"
		case CardYellowM1:
			return "yellow-1"
		case CardYellowM2:
			return "yellow-2"
		case CardYellow, CardYellowUK, CardYellowRN:
			return "yellow+?"
		default:
			return card.Short()
		}
	}
	return letter
}
