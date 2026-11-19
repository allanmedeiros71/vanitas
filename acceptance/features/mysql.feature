@integration
Feature: MySQL and MariaDB driver
  Vanitas connects to MySQL 8+ and MariaDB 10.6+ and reads their catalog and rows.

  @R-02 @T-02-05
  Scenario Outline: Connect and report the server
    Given the <server> seed server
    When I connect to the <server> server as "app" with database "app"
    Then the server product is "<product>"
    And the server version starts with "<major>"

    Examples:
      | server     | product | major |
      | MySQL 8    | MySQL   | 8.    |
      | MariaDB 11 | MariaDB | 11.   |

  @R-02 @T-02-05
  Scenario Outline: Without a database every accessible database is listed
    Given the <server> seed server
    When I connect to the <server> server as "app" with database ""
    Then the databases are "app, reports"
    And the schemas of "reports" are "reports"

    Examples:
      | server     |
      | MySQL 8    |
      | MariaDB 11 |

  @R-02 @T-02-05
  Scenario Outline: Objects of the app database
    Given the <server> seed server
    When I connect to the <server> server as "app" with database "app"
    Then the objects of schema "app" are:
      | kind     | name         |
      | table    | events       |
      | table    | order_items  |
      | table    | orders       |
      | table    | teams        |
      | table    | users        |
      | view     | active_users |
      | function | user_count   |
      | trigger  | orders_audit |

    Examples:
      | server     |
      | MySQL 8    |
      | MariaDB 11 |

  @R-02 @T-02-05
  Scenario Outline: Columns of a table
    Given the <server> seed server
    When I connect to the <server> server as "app" with database "app"
    Then the columns of "app.users" are:
      | name        | type          | nullable |
      | id          | int           | no       |
      | team_id     | int           | yes      |
      | email       | varchar(120)  | no       |
      | name        | varchar(80)   | no       |
      | balance     | decimal(12,2) | no       |
      | external_id | char(36)      | yes      |
      | profile     | <json_type>   | yes      |
      | avatar      | blob          | yes      |
      | created_at  | datetime      | no       |

    Examples:
      | server     | json_type |
      | MySQL 8    | json      |
      | MariaDB 11 | longtext  |

  @R-02 @T-02-05
  Scenario Outline: Keys indexes and constraints
    Given the <server> seed server
    When I connect to the <server> server as "app" with database "app"
    Then the primary key of "app.order_items" is "order_id, line_no"
    And the primary key of "app.events" is empty
    And the foreign keys of "app.orders" are "user_id -> users(id)"
    And the indexes of "app.orders" include "idx_orders_user_status" on "user_id, status"
    And the unique constraints of "app.users" include "email"
    And the check constraints of "app.orders" mention "status"
    And the estimated row count of "app.events" is known

    Examples:
      | server     |
      | MySQL 8    |
      | MariaDB 11 |

  @R-02 @T-02-05
  Scenario Outline: Values are shown as the server writes them
    Given the <server> seed server
    When I connect to the <server> server as "app" with database "app"
    And I query "SELECT balance, external_id, created_at, avatar, team_id FROM users WHERE id = 70"
    Then row 1 shows "87.50", "00000000-0000-4000-8000-000000000070", "2026-01-01 01:10:00", "<bytes 1.2 KB>" and "NULL"

    Examples:
      | server     |
      | MySQL 8    |
      | MariaDB 11 |

  @R-02 @R-11 @T-02-05
  Scenario Outline: A read-only connection rejects writes
    Given the <server> seed server
    When I connect read-only to the <server> server as "app" with database "app"
    And I execute "DELETE FROM teams"
    Then the error contains "1792"
    And "SELECT count(*) FROM teams" returns "5"

    Examples:
      | server     |
      | MySQL 8    |
      | MariaDB 11 |

  @R-02 @R-LGPD @T-02-05
  Scenario Outline: A wrong password is an authentication error without the password
    Given the <server> seed server
    When I connect to the <server> server as "app" with password "wrong-Pw-123"
    Then the error is an authentication error
    And the error contains "1045"
    And the error does not contain "wrong-Pw-123"

    Examples:
      | server     |
      | MySQL 8    |
      | MariaDB 11 |
