package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	chastisafe "github.com/zastraeaxyz/chastisafe-go"
)

const (
	envDeveloperCode = "CHASTISAFE_DEV_CODE"
	envBaseURL       = "CHASTISAFE_BASE_URL"
)

type globalOptions struct {
	dryRun bool
	yes    bool
}

// Execute runs the CLI.
func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	opts := &globalOptions{}
	root := &cobra.Command{
		Use:     "chastisafe",
		Short:   "Developer utility for the chastisafe-go package",
		Version: buildVersion(),
		Long: `A proof-of-concept CLI for the chastisafe-go package.

The developer code is read from CHASTISAFE_DEV_CODE. Write commands print a
dry run and ask for confirmation before changing anything.`,
		Example: `  chastisafe who
  chastisafe templates list
  chastisafe locks latest
  chastisafe locks load <template-id> --code 1234 --purpose CHASTITY`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if opts.yes && opts.dryRun {
				return errors.New("--yes and --dry-run cannot be used together")
			}
			return nil
		},
	}
	root.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", false, "print the write and exit without applying it")
	root.PersistentFlags().BoolVar(&opts.yes, "yes", false, "apply write commands without prompting")
	root.AddCommand(
		newWhoCmd(),
		newTemplatesCmd(opts),
		newLocksCmd(opts),
	)
	return root
}

func newWhoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "who",
		Short: "Show the account attached to the developer code",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			identity, err := client.Who(cmd.Context())
			return show(cmd, identity, err)
		},
	}
}

func newClient() (*chastisafe.Client, error) {
	code := os.Getenv(envDeveloperCode)
	if code == "" {
		return nil, fmt.Errorf("%s is not set", envDeveloperCode)
	}
	return chastisafe.New(chastisafe.Options{
		Token:   code,
		BaseURL: os.Getenv(envBaseURL),
	})
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

// approve prints a dry run of a write and asks whether to apply it.
func approve(cmd *cobra.Command, opts *globalOptions, label string, payload any) bool {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "dry run: "+label)
	if payload != nil {
		fmt.Fprintln(out, prettyJSON(payload))
	}
	if opts.dryRun {
		return false
	}
	if opts.yes {
		return true
	}

	fmt.Fprint(out, "Apply? [y/N] ")
	var answer string
	_, err := fmt.Fscan(cmd.InOrStdin(), &answer)
	if !isTerminal(cmd.InOrStdin()) {
		fmt.Fprintln(out)
	}
	if err != nil {
		fmt.Fprintln(out, "aborted")
		return false
	}
	if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
		fmt.Fprintln(out, "aborted")
		return false
	}
	return true
}

func isTerminal(r io.Reader) bool {
	file, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func prettyJSON(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func jsonObject(raw string) (map[string]any, error) {
	data := []byte(raw)
	if strings.HasPrefix(raw, "@") {
		path := strings.TrimPrefix(raw, "@")
		if path == "" {
			return nil, errors.New("file path after @ is empty")
		}
		file, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		data = file
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("argument must be a JSON object: %w", err)
	}
	if spec == nil {
		return nil, errors.New("argument must be a JSON object")
	}
	return spec, nil
}

func show(cmd *cobra.Command, value any, err error) error {
	if err != nil {
		return err
	}
	return writeJSON(cmd.OutOrStdout(), value)
}

func showLock(cmd *cobra.Command, lock *chastisafe.LoadedLock, err error) error {
	if err != nil {
		return err
	}
	if lock == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "ok")
		return nil
	}
	return writeJSON(cmd.OutOrStdout(), lock)
}

func done(cmd *cobra.Command, err error) error {
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok")
	return nil
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
