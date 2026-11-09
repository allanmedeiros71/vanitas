// Package acceptance runs the Gherkin scenarios against the product boundary.
package acceptance

import (
	"testing"

	"github.com/cucumber/godog"
)

// runFeature runs features/<name>.feature in strict mode.
func runFeature(t *testing.T, name string, init func(*godog.ScenarioContext)) {
	t.Helper()
	suite := godog.TestSuite{
		ScenarioInitializer: init,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/" + name + ".feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("feature %s failed", name)
	}
}
