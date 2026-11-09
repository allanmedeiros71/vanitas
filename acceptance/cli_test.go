package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/allanmedeiros71/vanitas/internal/cli"
)

type cliWorld struct {
	info   cli.BuildInfo
	stdout bytes.Buffer
	stderr bytes.Buffer
	code   int
}

func (w *cliWorld) build(version, commit, date string) {
	w.info = cli.BuildInfo{Version: version, Commit: commit, Date: date}
}

func (w *cliWorld) noBuildInfo() {
	w.info = cli.BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
}

func (w *cliWorld) run(line string) {
	args := strings.Fields(line)[1:]
	cmd := cli.NewRootCommand(w.info)
	cmd.SetOut(&w.stdout)
	cmd.SetErr(&w.stderr)
	cmd.SetArgs(args)
	w.code = 0
	if err := cmd.Execute(); err != nil {
		w.code = 1
	}
}

func (w *cliWorld) exitCode(want int) error {
	if w.code != want {
		return fmt.Errorf("exit code %d, want %d", w.code, want)
	}
	return nil
}

func (w *cliWorld) output(want string) error {
	if got := strings.TrimSuffix(w.stdout.String(), "\n"); got != want {
		return fmt.Errorf("output %q, want %q", got, want)
	}
	return nil
}

func (w *cliWorld) errorContains(want string) error {
	if !strings.Contains(w.stderr.String(), want) {
		return fmt.Errorf("error output %q does not contain %q", w.stderr.String(), want)
	}
	return nil
}

func TestCLIFeature(t *testing.T) {
	runFeature(t, "cli", func(sc *godog.ScenarioContext) {
		w := &cliWorld{}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			*w = cliWorld{}
			w.noBuildInfo()
			return ctx, nil
		})
		sc.Step(`^the build version "([^"]*)", commit "([^"]*)" and date "([^"]*)"$`, w.build)
		sc.Step(`^a build without version information$`, w.noBuildInfo)
		sc.Step(`^I run "([^"]*)"$`, w.run)
		sc.Step(`^the exit code is (\d+)$`, w.exitCode)
		sc.Step(`^the output is "([^"]*)"$`, w.output)
		sc.Step(`^the error output contains "((?:[^"\\]|\\.)*)"$`, func(s string) error {
			return w.errorContains(strings.ReplaceAll(s, `\"`, `"`))
		})
	})
}
