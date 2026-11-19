Feature: Passwords and the connections command
  Passwords come from the URL, password_command or the OS keyring and never reach files,
  errors or the screen.

  Background:
    Given an empty configuration directory
    And a working keyring

  @R-01 @R-13 @T-02-07
  Scenario: List saved connections
    Given the connections file:
      """
      [[connection]]
      name = "local-pg"
      driver = "postgres"
      host = "localhost"
      port = 5432
      user = "app"
      database = "app"

      [[connection]]
      name = "shop"
      driver = "mysql"
      host = "10.0.0.5"
      user = "report"
      read_only = true
      production = true
      """
    When I run "vanitas connections list"
    Then the exit code is 0
    And the listed connections are:
      | NAME     | DRIVER   | TARGET                 | FLAGS   |
      | local-pg | postgres | app@localhost:5432/app |         |
      | shop     | mysql    | report@10.0.0.5:3306   | RO,PROD |

  @R-01 @R-13 @T-02-07
  Scenario: List with no saved connections
    When I run "vanitas connections list"
    Then the exit code is 0
    And the output is "No saved connections. Add one with: vanitas connections add <name> <url>"

  @R-01 @R-13 @R-LGPD @T-02-07
  Scenario: Add a connection from a URL keeps the password in the keyring
    When I run "vanitas connections add local-pg postgres://app:s3cr3t-Pw@localhost:5432/app --read-only"
    Then the exit code is 0
    And the output is "saved local-pg"
    And the keyring holds "s3cr3t-Pw" for "local-pg"
    And the connections file contains "read_only = true"
    And the connections file does not contain "s3cr3t-Pw"

  @R-01 @R-13 @T-02-07
  Scenario: Add rejects a duplicate name
    Given a new saved connection "notes" for "./notes.db"
    When I run "vanitas connections add notes ./other.db"
    Then the exit code is 1
    And the error output contains "a connection named \"notes\" already exists"

  @R-01 @R-LGPD @T-02-07
  Scenario: Add without a keyring never stores the password in plaintext
    Given the keyring is unavailable
    When I run "vanitas connections add local-pg postgres://app:s3cr3t-Pw@localhost:5432/app"
    Then the exit code is 0
    And the error output contains "keyring unavailable: the password was not saved and will be asked when connecting"
    And the connections file does not contain "s3cr3t-Pw"

  @R-01 @R-13 @T-02-07
  Scenario: Remove a connection and its password
    Given a new saved connection "local-pg" for "postgres://app:s3cr3t-Pw@localhost:5432/app"
    When I run "vanitas connections remove local-pg"
    Then the exit code is 0
    And the output is "removed local-pg"
    And the keyring holds nothing for "local-pg"
    And the connections file does not contain "local-pg"

  @R-01 @R-13 @T-02-07
  Scenario: Remove an unknown connection
    When I run "vanitas connections remove nope"
    Then the exit code is 1
    And the error output contains "unknown connection \"nope\""

  @R-01 @T-02-07
  Scenario: The keyring password is used when connecting
    Given a new saved connection "local-pg" for "postgres://app@localhost:5432/app"
    And the keyring holds "s3cr3t-Pw" for "local-pg"
    When connection "local-pg" is resolved
    Then the resolved password is "s3cr3t-Pw"
    And the resolved configuration printed as text does not contain "s3cr3t-Pw"

  @R-01 @unix @T-02-07
  Scenario: password_command wins over the keyring
    Given a new saved connection "local-pg" for "postgres://app@localhost:5432/app" with password command "printf 'from-cmd-Pw\n'"
    And the keyring holds "s3cr3t-Pw" for "local-pg"
    When connection "local-pg" is resolved
    Then the resolved password is "from-cmd-Pw"

  @R-01 @R-LGPD @unix @T-02-07
  Scenario: A failing password_command does not leak its output
    Given a new saved connection "local-pg" for "postgres://app@localhost:5432/app" with password command "echo token-expired-XYZ >&2; echo partial-Pw; exit 3"
    When connection "local-pg" is resolved
    Then the error is "connection \"local-pg\": password_command failed: exit status 3"

  @R-01 @unix @T-02-07
  Scenario: A slow password_command times out
    Given a password_command timeout of 1 second
    And a new saved connection "local-pg" for "postgres://app@localhost:5432/app" with password command "sleep 5"
    When connection "local-pg" is resolved
    Then the error is "connection \"local-pg\": password_command timed out after 1s"
