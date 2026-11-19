Feature: SQLite driver
  Vanitas opens existing SQLite files and reads their catalog and rows.

  Background:
    Given a copy of the SQLite seed database named "seed.db"

  @R-02 @T-02-03
  Scenario: Open a SQLite file
    When I open the SQLite file "seed.db"
    Then the server product is "SQLite"
    And the server version is at least "3.35"
    And the databases are "main"
    And the schemas of "main" are "main"

  @R-02 @T-02-03
  Scenario: Objects of the main schema
    When I open the SQLite file "seed.db"
    Then the objects of schema "main" are:
      | kind    | name         |
      | table   | events       |
      | table   | order_items  |
      | table   | orders       |
      | table   | teams        |
      | table   | users        |
      | view    | active_users |
      | trigger | orders_audit |

  @R-02 @T-02-03
  Scenario: Columns of a table
    When I open the SQLite file "seed.db"
    Then the columns of "main.users" are:
      | name        | type          | nullable |
      | id          | INTEGER       | no       |
      | team_id     | INTEGER       | yes      |
      | email       | VARCHAR(120)  | no       |
      | name        | VARCHAR(80)   | no       |
      | balance     | DECIMAL(12,2) | no       |
      | external_id | TEXT          | yes      |
      | profile     | TEXT          | yes      |
      | avatar      | BLOB          | yes      |
      | created_at  | TIMESTAMP     | no       |

  @R-02 @T-02-03
  Scenario: Keys indexes and constraints
    When I open the SQLite file "seed.db"
    Then the primary key of "main.order_items" is "order_id, line_no"
    And the primary key of "main.events" is empty
    And the foreign keys of "main.orders" are "user_id -> users(id)"
    And the foreign keys of "main.users" are "team_id -> teams(id)"
    And the indexes of "main.users" include "idx_users_name" on "name"
    And the unique constraints of "main.users" include "email"
    And the check constraints of "main.orders" mention "status"

  @R-02 @T-02-03
  Scenario: Rows keep NULL and empty text and bytes apart
    When I open the SQLite file "seed.db"
    And I query "SELECT '' AS empty, NULL AS nothing, avatar FROM users WHERE id = 10"
    Then the result columns are "empty, nothing, avatar"
    And cell 1 of row 1 is the empty string
    And cell 2 of row 1 is NULL
    And cell 3 of row 1 shows "<bytes 1.2 KB>"

  @R-02 @T-02-03
  Scenario: Rows are read in batches
    When I open the SQLite file "seed.db"
    And I read "SELECT * FROM events" in batches of 4000 rows
    Then the batch sizes are "4000, 4000, 2000"

  @R-02 @R-11 @T-02-03
  Scenario: A read-only connection rejects writes
    When I open the SQLite file "seed.db" read-only
    And I execute "DELETE FROM teams"
    Then the error contains "readonly"
    And "SELECT count(*) FROM teams" returns "5"

  @R-02 @T-02-03
  Scenario: A missing file is not created
    When I open the SQLite file "missing.db"
    Then the error contains "sqlite: file not found:"
    And the file "missing.db" does not exist
