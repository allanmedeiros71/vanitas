package acceptance

import (
	"context"
	"io"
	"testing"

	"github.com/cucumber/godog"
)

// countScenarios runs config.feature with the given tag expression and counts the scenarios
// godog starts. Step definitions are not registered: the count only proves which scenarios the
// tag filter lets through, not whether they pass.
func countScenarios(t *testing.T, tags string) int {
	t.Helper()
	n := 0
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				n++
				return ctx, nil
			})
		},
		Options: &godog.Options{
			Format: "progress",
			Output: io.Discard,
			Paths:  []string{"features/config.feature"},
			Tags:   tags,
		},
	}
	_ = suite.Run()
	return n
}

func TestUnixTagExclusionRunsFewerScenarios(t *testing.T) {
	all := countScenarios(t, "")
	if all != 8 {
		t.Fatalf("without a tag filter config.feature starts %d scenarios, want 8", all)
	}
	withoutUnix := countScenarios(t, tagFilter("windows"))
	if withoutUnix != 6 {
		t.Errorf("with %q config.feature starts %d scenarios, want 6 (the two @unix ones excluded)", tagFilter("windows"), withoutUnix)
	}
}
