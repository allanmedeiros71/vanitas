// Package cli wires the vanitas command line.
package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Build information, set by ldflags (-X) at release time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// BuildInfo describes the running binary.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// NewRootCommand builds the vanitas root command.
func NewRootCommand(info BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:          "vanitas",
		Short:        "A terminal database client",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Placeholder until the TUI replaces it.
			return cmd.Help()
		},
	}
	root.AddCommand(newVersionCommand(info))
	return root
}

// Execute runs the CLI with the process arguments and returns the exit code.
func Execute() int {
	return execute(BuildInfo{Version: version, Commit: commit, Date: date}, os.Args[1:], os.Stdout, os.Stderr)
}

// execute is Execute with injectable arguments and streams.
func execute(info BuildInfo, args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand(info)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}
