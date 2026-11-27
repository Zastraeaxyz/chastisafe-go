package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	chastisafe "github.com/zastraeaxyz/chastisafe-go"
)

func newLocksLoadCmd(opts *globalOptions) *cobra.Command {
	var (
		code     string
		purpose  string
		trust    bool
		passcode string
		fakes    []string
	)
	cmd := &cobra.Command{
		Use:   "load <template-id>",
		Short: "Start a lock from a template",
		Long: `Start a lock from a template.

The lock code is required. The optional purpose is CHASTITY, BONDAGE, TASK,
LOYALTY, EMERGENCY or TEST.`,
		Example: "  chastisafe locks load <template-id> --code 1234 --purpose CHASTITY",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := chastisafe.LoadRequest{
				TemplateID: chastisafe.ID(args[0]),
				Code:       code,
				Trust:      trust,
				Fakes:      fakes,
			}
			if purpose != "" {
				parsed, err := chastisafe.ParsePurpose(purpose)
				if err != nil {
					return err
				}
				req.Purpose = parsed
			}
			if cmd.Flags().Changed("passcode") {
				req.Passcode = &passcode
			}

			payload := map[string]any{
				"templateId": args[0],
				"code":       "[redacted]",
				"purpose":    req.Purpose,
				"trust":      req.Trust,
				"fakes":      req.Fakes,
			}
			if req.Passcode != nil {
				payload["passcode"] = "[redacted]"
			}
			if !approve(cmd, opts, fmt.Sprintf("locks load %s", args[0]), payload) {
				return nil
			}

			client, err := newClient()
			if err != nil {
				return err
			}
			locks, err := client.Load(cmd.Context(), req)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), locks)
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "lock code (required)")
	cmd.Flags().StringVar(&purpose, "purpose", "", "lock purpose: CHASTITY, BONDAGE, TASK, LOYALTY, EMERGENCY or TEST")
	cmd.Flags().BoolVar(&trust, "trust", false, "start the lock with trust")
	cmd.Flags().StringVar(&passcode, "passcode", "", "optional passcode")
	cmd.Flags().StringSliceVar(&fakes, "fakes", nil, "fake lock codes")
	_ = cmd.MarkFlagRequired("code")
	_ = cmd.RegisterFlagCompletionFunc("purpose", completePurpose)
	return cmd
}

func newLocksOpenCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "open <lock-id>",
		Short: "Open a ready lock as the lockee",
		Long: `Open a ready lock as the lockee.

This calls the lockee endpoint /open/{id}. Use keyholder-unlock to act as the
keyholder instead.`,
		Example: "  chastisafe locks open <lock-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks open %s", args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			result, err := client.Open(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, result, err)
		},
	}
}

func newLocksKeyholderUnlockCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "keyholder-unlock <lock-id>",
		Short: "Unlock a lockee's lock as the keyholder",
		Long: `Unlock a lockee's lock as the keyholder.

This calls the keyholder endpoint /unlock/{id}. Use open to act as the lockee
instead.`,
		Example: "  chastisafe locks keyholder-unlock <lock-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks keyholder-unlock %s", args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			result, err := client.Unlock(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, result, err)
		},
	}
}

func newLocksResetCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "reset <lock-id>",
		Short:   "Reset a lock to the beginning",
		Example: "  chastisafe locks reset <lock-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks reset %s", args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			result, err := client.Reset(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, result, err)
		},
	}
}

func newLocksPickCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "pick <lock-id>",
		Short:   "Draw the next card from a card lock",
		Example: "  chastisafe locks pick <lock-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks pick %s", args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			result, err := client.Pick(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, result, err)
		},
	}
}

func newLocksAbandonCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "abandon <lock-id>",
		Short:   "End a lock without revealing its code",
		Example: "  chastisafe locks abandon <lock-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("locks abandon %s", args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			return done(cmd, client.Abandon(cmd.Context(), chastisafe.ID(args[0])))
		},
	}
}

func completePurpose(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(chastisafe.PurposeBondage),
		string(chastisafe.PurposeChastity),
		string(chastisafe.PurposeEmergency),
		string(chastisafe.PurposeLoyalty),
		string(chastisafe.PurposeTask),
		string(chastisafe.PurposeTest),
	}, cobra.ShellCompDirectiveNoFileComp
}
