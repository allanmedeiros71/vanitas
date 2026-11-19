Feature: Connection session
  vanitas connects from the command line, shows the connection in the title bar, asks for a
  missing password and disconnects.

  Background:
    Given an empty configuration directory
    And a working keyring
    And a copy of the SQLite seed database named "seed.db"
    And a terminal of 120x40

  @R-01 @R-13 @T-02-08
  Scenario: Open a saved connection with -c
    Given a new saved connection "notes" for the SQLite file "seed.db"
    When I start vanitas with "-c notes"
    And the connection finishes
    Then the screen shows "vanitas ─ conn: notes ─ db: seed.db"
    And the status bar contains "connected to notes · SQLite 3."
    And the explorer shows "SQLite 3."
    And the screen matches the golden file "session_connected"

  @R-01 @R-13 @T-02-08
  Scenario: Open a SQLite file from the command line
    When I start vanitas with the path of the SQLite file "seed.db"
    And the connection finishes
    Then the screen shows "─ db: seed.db"
    And the status bar contains "· SQLite 3."

  @R-01 @R-11 @T-02-08
  Scenario: Read-only and production tags in the title bar
    Given a new saved connection "notes" for the SQLite file "seed.db" marked read-only and production
    When I start vanitas with "-c notes"
    And the connection finishes
    Then the screen shows "vanitas ─ conn: notes (RO, PROD) ─ db: seed.db"

  @R-13 @T-02-08
  Scenario: Unknown saved connection
    When I run "vanitas -c nope"
    Then the exit code is 1
    And the error output contains "unknown connection \"nope\""

  @R-13 @T-02-08
  Scenario: Malformed target
    When I run "vanitas oracle://x"
    Then the exit code is 1
    And the error output contains "unsupported URL scheme \"oracle\""

  @R-01 @T-02-08
  Scenario: Disconnect from the leader menu
    Given a new saved connection "notes" for the SQLite file "seed.db"
    When I start vanitas with "-c notes"
    And the connection finishes
    And I press "space"
    Then the popup lists "x" with "disconnect"
    When I press "x"
    And the disconnection finishes
    Then the screen shows "vanitas ─ no connection"
    And the status bar contains "disconnected from notes"

  @R-01 @T-02-08
  Scenario: Disconnect without a connection
    When vanitas starts
    And I press "space"
    And I press "x"
    Then the status bar contains "not connected"

  @R-01 @R-LGPD @T-02-08
  Scenario: Debug logs and the screen never show the URL password
    Given the environment variable "VANITAS_LOG" is "debug"
    When I start vanitas with "postgres://app:s3cr3t-Pw@127.0.0.1:1/app"
    And the connection finishes
    Then the screen does not contain "s3cr3t-Pw"
    And the log file does not contain "s3cr3t-Pw"

  @R-01 @R-LGPD @integration @T-02-08
  Scenario: A URL password is masked in the title bar
    Given the PostgreSQL 16 seed server
    When I start vanitas with the PostgreSQL URL "postgres://app:app-Pw-123@{host}:{port}/app"
    And the connection finishes
    Then the screen shows "conn: postgres://app:***@{host}:{port}/app ─ db: app"
    And the screen does not contain "app-Pw-123"

  @R-01 @R-LGPD @integration @T-02-08
  Scenario: A missing password is asked and not stored
    Given the PostgreSQL 16 seed server
    And a saved connection "local-pg" to the PostgreSQL server as "app" without password
    When I start vanitas with "-c local-pg"
    And the connection finishes
    Then the screen shows "Password for local-pg"
    When I type "app-Pw-123"
    Then the screen does not contain "app-Pw-123"
    When I press "enter"
    And the connection finishes
    Then the screen shows "vanitas ─ conn: local-pg ─ db: app"
    And the keyring holds nothing for "local-pg"
