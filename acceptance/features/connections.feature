Feature: Saved connections and connection targets
  Connections live in connections.toml without passwords; URLs and SQLite files open
  temporary connections; ${VAR} templates expand when connecting.

  Background:
    Given an empty configuration directory

  @R-01 @T-02-06
  Scenario: Load saved connections
    Given the connections file:
      """
      [[connection]]
      name = "local-pg"
      driver = "postgres"
      host = "localhost"
      port = 5432
      user = "app"
      database = "app"
      sslmode = "disable"

      [[connection]]
      name = "notes"
      driver = "sqlite"
      path = "~/data/notes.db"
      read_only = true
      """
    When the connections are loaded
    Then there are 2 connections
    And connection "local-pg" has driver "postgres" and target "app@localhost:5432/app"
    And connection "notes" has driver "sqlite" and target "~/data/notes.db"
    And connection "notes" is read-only

  @R-01 @R-LGPD @T-02-06
  Scenario Outline: Parse a connection target
    When I parse the target "<target>"
    Then the target has driver "<driver>" and target "<shown>"
    And the target is named "<name>"

    Examples:
      | target                                                             | driver   | shown                         | name                                                          |
      | postgres://app:s3cr3t-Pw@db.example.test:5433/app?sslmode=require  | postgres | app@db.example.test:5433/app  | postgres://app:***@db.example.test:5433/app?sslmode=require   |
      | postgresql://app@localhost/app                                     | postgres | app@localhost:5432/app        | postgresql://app@localhost/app                                |
      | mysql://root:pw-123@127.0.0.1:3307/shop                            | mysql    | root@127.0.0.1:3307/shop      | mysql://root:***@127.0.0.1:3307/shop                          |
      | mariadb://root@localhost/shop                                      | mysql    | root@localhost:3306/shop      | mariadb://root@localhost/shop                                 |
      | sqlite:///var/data/app.db                                          | sqlite   | /var/data/app.db              | sqlite:///var/data/app.db                                     |
      | ./notes.sqlite3                                                    | sqlite   | ./notes.sqlite3               | ./notes.sqlite3                                               |

  @R-01 @T-02-06
  Scenario Outline: Reject an invalid target
    When I parse the target "<target>"
    Then the error is "<error>"

    Examples:
      | target      | error                                                                       |
      | oracle://x  | unsupported URL scheme "oracle" (use postgres, mysql, mariadb or sqlite)     |
      | ./notes.txt | "./notes.txt" is not a URL or a SQLite file (.db, .sqlite, .sqlite3)        |

  @R-01 @R-LGPD @T-02-06
  Scenario: Environment variables expand when connecting and not in the file
    Given the environment variable "PROD_USER" is "report"
    And the environment variable "PROD_HOST" is not set
    And the connections file:
      """
      [[connection]]
      name = "prod"
      driver = "mysql"
      host = "${PROD_HOST:-10.0.0.5}"
      user = "${PROD_USER}"
      database = "shop"
      """
    When connection "prod" is resolved
    Then the resolved host is "10.0.0.5"
    And the resolved user is "report"
    And the resolved port is 3306
    When the connections are saved again
    Then the connections file contains "user = \"${PROD_USER}\""

  @R-01 @T-02-06
  Scenario: An unset variable without a default is an error
    Given the environment variable "PROD_USER" is not set
    And the connections file:
      """
      [[connection]]
      name = "prod"
      driver = "mysql"
      host = "10.0.0.5"
      user = "${PROD_USER}"
      """
    When connection "prod" is resolved
    Then the error is "connection \"prod\": environment variable PROD_USER is not set"

  @R-01 @T-02-06
  Scenario: Invalid connections file names the file and line
    Given the connections file:
      """
      [[connection]]
      name = "local-pg
      """
    When the connections are loaded
    Then the error starts with "connections.toml: line 2"

  @R-01 @T-02-06
  Scenario: Duplicate connection names are rejected
    Given the connections file:
      """
      [[connection]]
      name = "local-pg"
      driver = "postgres"
      host = "localhost"

      [[connection]]
      name = "local-pg"
      driver = "postgres"
      host = "other"
      """
    When the connections are loaded
    Then the error is "connections.toml: duplicate connection name \"local-pg\""

  @R-01 @R-16 @unix @T-02-06
  Scenario: Saving writes an owner-only file
    Given a new saved connection "notes" for "./notes.db"
    Then the file "connections.toml" has mode "0600"
    And the connections file contains "name = \"notes\""

  @R-01 @R-LGPD @T-02-06
  Scenario: Errors shown to the user have the password masked
    Given a connection error "pq: password authentication failed: s3cr3t-Pw at postgres://app:s3cr3t-Pw@localhost/app"
    When the error is redacted for the password "s3cr3t-Pw"
    Then the redacted error is "pq: password authentication failed: *** at postgres://app:***@localhost/app"
