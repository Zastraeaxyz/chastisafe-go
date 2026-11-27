package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	chastisafe "github.com/zastraeaxyz/chastisafe-go"
)

func newTemplatesCmd(opts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template"},
		Short:   "Inspect and manage lock templates",
		Long: `Inspect and manage lock templates.

Template IDs are returned by templates list and templates get. Use templates
edit to fetch the private edit shape before templates update.`,
	}
	cmd.AddCommand(
		newTemplatesListCmd(),
		newTemplatesGetCmd(),
		newTemplatesEditCmd(),
		newTemplatesCreateCmd(opts),
		newTemplatesUpdateCmd(opts),
		newTemplatesDeleteCmd(opts),
		newTemplatesSimulateCmd(),
	)
	return cmd
}

func newTemplatesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List templates",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			templates, err := client.Templates(cmd.Context())
			return show(cmd, templates, err)
		},
	}
}

func newTemplatesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <template-id>",
		Aliases: []string{"show"},
		Short:   "Show one template",
		Example: "  chastisafe templates get 49d9927e-bf50-0756-9ae3-73289dcc84b5",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			template, err := client.Template(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, template, err)
		},
	}
}

func newTemplatesEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <template-id>",
		Short: "Show the private edit shape of a template",
		Long: `Show the private edit shape of a template.

Copy and modify this output, then pass it to templates update.`,
		Example: "  chastisafe templates edit 49d9927e-bf50-0756-9ae3-73289dcc84b5",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			template, err := client.TemplateForEdit(cmd.Context(), chastisafe.ID(args[0]))
			return show(cmd, template, err)
		},
	}
}

func newTemplatesCreateCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "create <json|@file>",
		Aliases: []string{"new"},
		Short:   "Create a template from a JSON payload",
		Long: `Create a template from a JSON payload.

The payload is passed to POST /create. Pass inline JSON, or use @path to read
it from a file. Use templates list to find the new ID.`,
		Example: `  chastisafe templates create '{"lockname":"Demo","locktype":"CARDLOCK","cards":[{"name":"GREEN","min":1,"max":1}]}'
  chastisafe templates create @template.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := jsonObject(args[0])
			if err != nil {
				return err
			}
			if !approve(cmd, opts, "templates create", spec) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			return done(cmd, client.CreateTemplate(cmd.Context(), spec))
		},
	}
}

func newTemplatesUpdateCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "update <template-id> <json|@file>",
		Short: "Update a template from a JSON payload",
		Long: `Update a template from a JSON payload.

Pass inline JSON, or use @path to read it from a file. The template ID is added
to the payload automatically. Use templates edit to fetch the current shape and
templates get to read the result.`,
		Example: `  chastisafe templates update <template-id> '{"description":"Updated"}'
  chastisafe templates update <template-id> @template.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := jsonObject(args[1])
			if err != nil {
				return err
			}
			spec["templateId"] = chastisafe.ID(args[0])
			delete(spec, "id")
			if !approve(cmd, opts, fmt.Sprintf("templates update %s", args[0]), spec) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			return done(cmd, client.UpdateTemplate(cmd.Context(), spec))
		},
	}
}

func newTemplatesDeleteCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <template-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a template",
		Example: "  chastisafe templates delete 49d9927e-bf50-0756-9ae3-73289dcc84b5",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve(cmd, opts, fmt.Sprintf("templates delete %s", args[0]), nil) {
				return nil
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			return done(cmd, client.DeleteTemplate(cmd.Context(), chastisafe.ID(args[0])))
		},
	}
}

func newTemplatesSimulateCmd() *cobra.Command {
	var runs int
	cmd := &cobra.Command{
		Use:   "simulate <template-id>",
		Short: "Preview a card template's duration and card counts",
		Long: `Preview a card template's duration and card counts.

The result is read-only. The keyholder can disable simulation for a template,
in which case the command reports that and exits successfully.`,
		Example: "  chastisafe templates simulate 49d9927e-bf50-0756-9ae3-73289dcc84b5 --runs 3",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runs < 1 {
				return errors.New("--runs must be a positive integer")
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			templateID := chastisafe.ID(args[0])
			tmpl, err := client.Template(cmd.Context(), templateID)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s (%s)\n", tmpl.LockName, tmpl.LockType)
			for i := 1; i <= runs; i++ {
				sim, err := client.Simulate(cmd.Context(), templateID)
				if errors.Is(err, chastisafe.ErrSimulationDisabled) {
					fmt.Fprintln(out, "simulation is disabled for this lock")
					return nil
				}
				if err != nil {
					return fmt.Errorf("run %d: %w", i, err)
				}
				fmt.Fprintf(
					out,
					"run %d: avg %s (min %s, max %s) | cards avg %d (min %d, max %d) | n=%d\n",
					i,
					chastisafe.FormatDuration(time.Duration(sim.AvgDuration)*time.Second),
					chastisafe.FormatDuration(time.Duration(sim.MinDuration)*time.Second),
					chastisafe.FormatDuration(time.Duration(sim.MaxDuration)*time.Second),
					sim.AvgCards, sim.MinCards, sim.MaxCards,
					sim.NumberOfSimulations,
				)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&runs, "runs", 3, "number of simulations to run")
	return cmd
}
