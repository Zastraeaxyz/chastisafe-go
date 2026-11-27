package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	chastisafe "github.com/zastraeaxyz/chastisafe-go"
)

func newLocksCmd(opts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "locks",
		Aliases: []string{"lock"},
		Short:   "Inspect and operate loaded locks",
		Long: `Inspect and operate loaded locks.

Read commands run immediately. Commands that change a lock print a dry run and
ask for confirmation first. Use locks open for the lockee path and locks
keyholder-unlock for the keyholder path.`,
	}
	cmd.AddCommand(
		newLocksListCmd(),
		newLocksGetCmd(),
		newLocksRecentCmd(),
		newLocksLatestCmd(),
		newLocksLogCmd(),

		newLocksLoadCmd(opts),
		newLocksOpenCmd(opts),
		newLocksKeyholderUnlockCmd(opts),
		newLocksResetCmd(opts),
		newLocksPickCmd(opts),
		newLocksAbandonCmd(opts),

		newLocksAddTimeCmd(opts),
		newLocksSetRemainingCmd(opts),
		newLocksSetPurposeCmd(opts),
		newLocksFreezeCmd(opts),
		newLocksUnfreezeCmd(opts),
		newLocksHideCmd(opts),
		newLocksRevealCmd(opts),
		newLocksCardsCmd(opts),
	)
	return cmd
}

func newLocksListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active locks",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			locks, err := client.Loaded(cmd.Context())
			return show(cmd, locks, err)
		},
	}
}

func newLocksGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <lock-id>",
		Aliases: []string{"show"},
		Short:   "Show one active lock",
		Long: `Show one active lock.

The response also reports whether the caller is the lockee.`,
		Example: "  chastisafe locks get 49d9927e-bf50-0756-9ae3-73289dcc84b5",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			lock, _, err := client.LoadedLock(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, lock, err)
		},
	}
}

func newLocksRecentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recent",
		Short: "List completed locks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			recent, err := client.Recent(cmd.Context())
			return show(cmd, recent, err)
		},
	}
}

func newLocksLatestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "latest",
		Short: "Print the most recent lock ID",
		Long: `Print the most recent lock ID.

An active lock is preferred. When nothing is loaded, the newest completed lock
is used.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			id, err := latestLockID(cmd.Context(), client)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
}

func newLocksLogCmd() *cobra.Command {
	var from, count int
	cmd := &cobra.Command{
		Use:     "log <lock-id>",
		Short:   "Show lock history entries",
		Example: "  chastisafe locks log <lock-id> --from 0 --count 20",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			entries, err := client.Log(cmd.Context(), chastisafe.ID(args[0]), from, count)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), entries)
		},
	}
	cmd.Flags().IntVar(&from, "from", 0, "start offset")
	cmd.Flags().IntVar(&count, "count", 100, "maximum number of entries")
	return cmd
}

// latestLockID prefers an active lock, falling back to the most recently
// completed one when nothing is loaded.
func latestLockID(ctx context.Context, client *chastisafe.Client) (chastisafe.ID, error) {
	locks, err := client.Loaded(ctx)
	if err != nil {
		return "", err
	}
	if len(locks) > 0 {
		newest := locks[0]
		for _, lock := range locks[1:] {
			if lock.LoadTime.Time.After(newest.LoadTime.Time) {
				newest = lock
			}
		}
		return newest.ID, nil
	}

	history, err := client.Recent(ctx)
	if err != nil {
		return "", err
	}
	if len(history) == 0 {
		return "", errors.New("no loaded or completed locks")
	}
	newest := history[0]
	for _, lock := range history[1:] {
		if lock.UnlockTime.Time.After(newest.UnlockTime.Time) {
			newest = lock
		}
	}
	return newest.LockID, nil
}
