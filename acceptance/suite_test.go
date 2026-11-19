// Package acceptance runs the Gherkin scenarios against the product boundary.
package acceptance

import (
	"runtime"
	"testing"

	"github.com/cucumber/godog"
)

// tagFilter returns the godog tag expression for goos: @unix scenarios assume Unix paths
// or file modes, so they are excluded on Windows (D-20).
func tagFilter(goos string) string {
	if goos == "windows" {
		return "~@unix"
	}
	return ""
}

func TestTagFilter(t *testing.T) {
	for goos, want := range map[string]string{"windows": "~@unix", "linux": "", "darwin": ""} {
		if got := tagFilter(goos); got != want {
			t.Errorf("tagFilter(%q) = %q, want %q", goos, got, want)
		}
	}
}

// runFeature runs features/<name>.feature in strict mode.
func runFeature(t *testing.T, name string, init func(*godog.ScenarioContext)) {
	t.Helper()
	suite := godog.TestSuite{
		ScenarioInitializer: init,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/" + name + ".feature"},
			Strict:   true,
			Tags:     tagFilter(runtime.GOOS),
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("feature %s failed", name)
	}
}
