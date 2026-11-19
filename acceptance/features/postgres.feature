@integration
Feature: PostgreSQL driver
  Vanitas connects to PostgreSQL 12+ and reads its catalog and rows (seed in testdata/seed/).

  Background:
    Given the PostgreSQL 16 seed server

  @R-02 @T-02-04
  Scenario: Connect and report the server
    When I connect to the PostgreSQL server as "app" with database "app"
    Then the server product is "PostgreSQL"
    And the server version starts with "16."

  @R-02 @T-02-04
  Scenario: Without a database every accessible database is listed
    When I connect to the PostgreSQL server as "app" with database ""
    Then the databases include "app", "postgres" and "reports"
    And the databases do not include "template0" or "template1"

  @R-02 @T-02-04
  Scenario: System schemas are hidden
    When I connect to the PostgreSQL server as "app" with database "app"
    Then the schemas of "app" are "audit, public"
    And the schemas of "reports" are "public"

  @R-02 @T-02-04
  Scenario: Objects of the public schema
    When I connect to the PostgreSQL server as "app" with database "app"
    Then the objects of schema "public" are:
      | kind     | name            |
      | table    | order_items     |
      | table    | orders          |
      | table    | teams           |
      | table    | users           |
      | view     | active_users    |
      | function | orders_audit_fn |
      | function | user_count      |
      | sequence | invoice_seq     |
      | trigger  | orders_audit    |
    And the objects of schema "audit" are:
      | kind  | name   |
      | table | events |

  @R-02 @T-02-04
  Scenario: Columns of a table
    When I connect to the PostgreSQL server as "app" with database "app"
    Then the columns of "public.users" are:
      | name        | type                        | nullable |
      | id          | integer                     | no       |
      | team_id     | integer                     | yes      |
      | email       | character varying(120)      | no       |
      | name        | character varying(80)       | no       |
      | balance     | numeric(12,2)               | no       |
      | external_id | uuid                        | yes      |
      | profile     | jsonb                       | yes      |
      | avatar      | bytea                       | yes      |
      | created_at  | timestamp without time zone | no       |

  @R-02 @T-02-04
  Scenario: Keys indexes and constraints
    When I connect to the PostgreSQL server as "app" with database "app"
    Then the primary key of "public.order_items" is "order_id, line_no"
    And the primary key of "audit.events" is empty
    And the foreign keys of "public.orders" are "user_id -> users(id)"
    And the indexes of "public.orders" include "idx_orders_user_status" on "user_id, status"
    And the unique constraints of "public.users" include "email"
    And the check constraints of "public.orders" mention "status"
    And the estimated row count of "audit.events" is known

  @R-02 @T-02-04
  Scenario: Values are shown as the server writes them
    When I connect to the PostgreSQL server as "app" with database "app"
    And I query "SELECT balance, external_id, profile, created_at, avatar, team_id FROM users WHERE id = 70"
    Then row 1 shows "87.50", "00000000-0000-4000-8000-000000000070", "{\"n\": 70, \"theme\": \"dark\"}", "2026-01-01 01:10:00", "<bytes 1.2 KB>" and "NULL"

  @R-02 @R-11 @T-02-04
  Scenario: A read-only connection rejects writes
    When I connect read-only to the PostgreSQL server as "app" with database "app"
    And I execute "DELETE FROM teams"
    Then the error contains "25006"
    And "SELECT count(*) FROM teams" returns "5"

  @R-02 @R-LGPD @T-02-04
  Scenario: A wrong password is an authentication error without the password
    When I connect to the PostgreSQL server as "app" with password "wrong-Pw-123"
    Then the error is an authentication error
    And the error does not contain "wrong-Pw-123"
