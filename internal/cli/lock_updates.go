package cli

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	chastisafe "github.com/zastraeaxyz/chastisafe-go"
)

func newLocksAddTimeCmd(opts *globalOptions) *cobra.Command {
	var raw string
	cmd := &cobra.Command{
		Use:   "add-time <lock-id>",
		Short: "Add or remove timer time",
		Long:  "Add or remove timer time. Negative durations remove time.",
		Example: `  chastisafe locks add-time <lock-id> --duration 1h30m
  chastisafe locks add-time <lock-id> --duration -10m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration, err := chastisafe.ParseDuration(raw)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"id":       args[0],
				"duration": duration.String(),
				"seconds":  int(duration / time.Second),
			}
			if !approve(cmd, opts, fmt.Sprintf("locks add-time %s", args[0]), payload) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, err := client.AddTime(cmd.Context(), chastisafe.ID(args[0]), duration)
			return showLock(cmd, lock, err)
		},
	}
	cmd.Flags().StringVar(&raw, "duration", "", "duration to add, for example 1h30m or -10m (required)")
	_ = cmd.MarkFlagRequired("duration")
	return cmd
}

func newLocksSetRemainingCmd(opts *globalOptions) *cobra.Command {
	var raw string
	cmd := &cobra.Command{
		Use:     "set-remaining <lock-id>",
		Short:   "Replace a timer's remaining time as the keyholder",
		Example: "  chastisafe locks set-remaining <lock-id> --duration 3d",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration, err := chastisafe.ParseDuration(raw)
			if err != nil {
				return err
			}
			if duration < 0 {
				return errors.New("--duration cannot be negative")
			}
			payload := map[string]any{
				"id":       args[0],
				"duration": duration.String(),
				"seconds":  int(duration / time.Second),
			}
			if !approve(cmd, opts, fmt.Sprintf("locks set-remaining %s", args[0]), payload) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, err := client.SetRemainingTime(cmd.Context(), chastisafe.ID(args[0]), duration)
			return showLock(cmd, lock, err)
		},
	}
	cmd.Flags().StringVar(&raw, "duration", "", "remaining time, for example 90m or 3d (required)")
	_ = cmd.MarkFlagRequired("duration")
	return cmd
}

func newLocksSetPurposeCmd(opts *globalOptions) *cobra.Command {
	var purpose string
	cmd := &cobra.Command{
		Use:     "set-purpose <lock-id>",
		Short:   "Change a lock's purpose",
		Example: "  chastisafe locks set-purpose <lock-id> --purpose CHASTITY",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := chastisafe.ParsePurpose(purpose)
			if err != nil {
				return err
			}
			payload := map[string]any{"id": args[0], "purpose": parsed}
			if !approve(cmd, opts, fmt.Sprintf("locks set-purpose %s", args[0]), payload) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, err := client.SetPurpose(cmd.Context(), chastisafe.ID(args[0]), parsed)
			return showLock(cmd, lock, err)
		},
	}
	cmd.Flags().StringVar(&purpose, "purpose", "", "lock purpose: CHASTITY, BONDAGE, TASK, LOYALTY, EMERGENCY or TEST (required)")
	_ = cmd.MarkFlagRequired("purpose")
	_ = cmd.RegisterFlagCompletionFunc("purpose", completePurpose)
	return cmd
}

func newLocksFreezeCmd(opts *globalOptions) *cobra.Command {
	return newLocksFreezeLikeCmd(opts, true, "freeze", "Freeze a lock as the keyholder")
}

func newLocksUnfreezeCmd(opts *globalOptions) *cobra.Command {
	return newLocksFreezeLikeCmd(opts, false, "unfreeze", "Unfreeze a lock as the keyholder")
}

func newLocksFreezeLikeCmd(opts *globalOptions, frozen bool, use, short string) *cobra.Command {
	return &cobra.Command{
		Use:     use + " <lock-id>",
		Short:   short,
		Example: fmt.Sprintf("  chastisafe locks %s <lock-id>", use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks %s %s", use, args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, err := client.Freeze(cmd.Context(), chastisafe.ID(args[0]), frozen)
			return showLock(cmd, lock, err)
		},
	}
}

func newLocksHideCmd(opts *globalOptions) *cobra.Command {
	return newLocksHideLikeCmd(opts, true, "hide", "Hide a lock as the keyholder")
}

func newLocksRevealCmd(opts *globalOptions) *cobra.Command {
	return newLocksHideLikeCmd(opts, false, "reveal", "Reveal a lock as the keyholder")
}

func newLocksHideLikeCmd(opts *globalOptions, hidden bool, use, short string) *cobra.Command {
	return &cobra.Command{
		Use:     use + " <lock-id>",
		Short:   short,
		Example: fmt.Sprintf("  chastisafe locks %s <lock-id>", use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks %s %s", use, args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, err := client.Hide(cmd.Context(), chastisafe.ID(args[0]), hidden)
			return showLock(cmd, lock, err)
		},
	}
}

func newLocksCardsCmd(opts *globalOptions) *cobra.Command {
	var add, remove, message string
	var show bool
	cmd := &cobra.Command{
		Use:   "cards <lock-id>",
		Short: "Add or remove cards on a card lock",
		Long: `Add or remove cards on a card lock.

Use name:count lists. A comma separates multiple cards.`,
		Example: `  chastisafe locks cards <lock-id> --add green:2,red:1
  chastisafe locks cards <lock-id> --remove yellow:1 --message "rebalanced" --show`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if add == "" && remove == "" {
				return errors.New("use --add or --remove")
			}

			var changes []chastisafe.CardChange
			if add != "" {
				cards, err := chastisafe.ParseCards(add)
				if err != nil {
					return err
				}
				for _, card := range sortedCards(cards) {
					changes = append(changes, chastisafe.CardChange{Card: card, Delta: cards[card]})
				}
			}
			if remove != "" {
				cards, err := chastisafe.ParseCards(remove)
				if err != nil {
					return err
				}
				for _, card := range sortedCards(cards) {
					changes = append(changes, chastisafe.CardChange{Card: card, Delta: -cards[card]})
				}
			}

			payload := map[string]any{
				"id":      args[0],
				"changes": changes,
				"message": message,
				"show":    show,
			}
			if !approve(cmd, opts, fmt.Sprintf("locks cards %s", args[0]), payload) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, err := client.UpdateCards(cmd.Context(), chastisafe.ID(args[0]), changes, chastisafe.UpdateOptions{
				Message: message,
				Show:    show,
			})
			return showLock(cmd, lock, err)
		},
	}
	cmd.Flags().StringVar(&add, "add", "", "cards to add, for example green:2,red:1")
	cmd.Flags().StringVar(&remove, "remove", "", "cards to remove, for example yellow:1")
	cmd.Flags().StringVar(&message, "message", "", "optional update message")
	cmd.Flags().BoolVar(&show, "show", false, "show the update in the lock log")
	_ = cmd.RegisterFlagCompletionFunc("add", completeCards)
	_ = cmd.RegisterFlagCompletionFunc("remove", completeCards)
	return cmd
}

func completeCards(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	prefix := ""
	if index := strings.LastIndex(toComplete, ","); index >= 0 {
		prefix = toComplete[:index+1]
	}
	names := cardNames()
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, prefix+name+":")
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func cardNames() []string {
	cards := []chastisafe.Card{
		chastisafe.CardGreen,
		chastisafe.CardRed,
		chastisafe.CardFreeze,
		chastisafe.CardArctic,
		chastisafe.CardStun,
		chastisafe.CardDice,
		chastisafe.CardPink,
		chastisafe.CardGoAgain,
		chastisafe.CardSticky,
		chastisafe.CardDoubleUp,
		chastisafe.CardReset,
		chastisafe.CardYellow,
		chastisafe.CardYellowP1,
		chastisafe.CardYellowP2,
		chastisafe.CardYellowP3,
		chastisafe.CardYellowM1,
		chastisafe.CardYellowM2,
		chastisafe.CardYellowUK,
		chastisafe.CardYellowRN,
	}
	names := make([]string, len(cards))
	for i, card := range cards {
		names[i] = strings.ToLower(string(card))
	}
	return names
}

func sortedCards(cards map[chastisafe.Card]int) []chastisafe.Card {
	list := make([]chastisafe.Card, 0, len(cards))
	for card := range cards {
		list = append(list, card)
	}
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
	return list
}
