Feature: Command line basics
  The vanitas binary reports its build and rejects unknown commands.

  @R-13 @T-01-01
  Scenario: Version with build information
    Given the build version "0.1.0", commit "abc1234" and date "2026-10-25"
    When I run "vanitas version"
    Then the exit code is 0
    And the output is "vanitas 0.1.0 (abc1234, 2026-10-25)"

  @R-13 @T-01-01
  Scenario: Version without build information
    Given a build without version information
    When I run "vanitas version"
    Then the exit code is 0
    And the output is "vanitas dev (none, unknown)"

  @R-13 @T-01-01
  Scenario: Unknown command
    When I run "vanitas frobnicate"
    Then the exit code is 1
    And the error output contains "unknown command \"frobnicate\""
