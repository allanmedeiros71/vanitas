Feature: Connection picker
  The picker lists, filters, tests, connects to, adds, edits and deletes saved connections.

  Background:
    Given an empty configuration directory
    And a working keyring
    And a copy of the SQLite seed database named "seed.db"
    And a terminal of 120x40
    And the saved connections:
      | name     | driver   | target                     | flags   | color | password  |
      | notes    | sqlite   | seed.db                    |         |       |           |
      | local-pg | postgres | app@127.0.0.1:1/app        |         |       | s3cr3t-Pw |
      | shop     | mysql    | report@127.0.0.1:1/shop    | RO,PROD | #dd1144 |           |

  @R-01 @T-02-09
  Scenario: The picker opens at startup with the saved connections
    When I start vanitas without arguments
    Then the screen shows "Connections"
    And the picker lists "notes", "local-pg" and "shop" in this order
    And the row "local-pg" shows "postgres" and "app@127.0.0.1:1/app"
    And the row "shop" shows "RO" and "PROD"
    And the footer contains "enter connect"

  @R-01 @T-02-09
  Scenario: Empty picker
    Given no saved connections
    When I start vanitas without arguments
    Then the screen shows "No connections yet · press a to add one"

  @R-01 @T-02-09
  Scenario: Filter the connections
    When I start vanitas without arguments
    And I press "/"
    And I type "sh"
    Then the picker lists "shop"
    When I press "esc"
    Then the picker lists "notes", "local-pg" and "shop" in this order
    When I press "/"
    And I type "zzz"
    Then the screen shows "No matching connections"

  @R-01 @T-02-09
  Scenario: Connect from the picker
    When I start vanitas without arguments
    And I press "enter"
    And the connection finishes
    Then the screen does not show "Connections"
    And the screen shows "vanitas ─ conn: notes ─ db: seed.db"

  @R-01 @R-LGPD @T-02-09
  Scenario: A failed connection keeps the user in the picker
    When I start vanitas without arguments
    And I press "down"
    And I press "enter"
    And the connection finishes
    Then the screen shows "Connections"
    And the status bar contains "could not connect to local-pg:"
    And the screen does not contain "s3cr3t-Pw"

  @R-01 @T-02-09
  Scenario: A failed connection from the command line opens the picker
    When I start vanitas with "./missing.db"
    And the connection finishes
    Then the screen shows "Connections"
    And the status bar contains "could not connect to ./missing.db: sqlite: file not found:"

  @R-01 @T-02-09
  Scenario: Test connections
    When I start vanitas without arguments
    And I press "t"
    And the test finishes
    Then the status bar contains "connection ok: SQLite 3."
    When I press "down"
    And I press "t"
    And the test finishes
    Then the status bar contains "connection failed:"
    And the screen does not contain "s3cr3t-Pw"
    And the screen shows "Connections"

  @R-01 @R-LGPD @T-02-09
  Scenario: Delete a connection after confirmation
    When I start vanitas without arguments
    And I press "down"
    And I press "d"
    Then the screen shows "Delete connection \"local-pg\"?"
    When I press "n"
    Then the picker lists "notes", "local-pg" and "shop" in this order
    When I press "d"
    And I press "y"
    And the save finishes
    Then the picker lists "notes" and "shop" in this order
    And the status bar contains "deleted local-pg"
    And the connections file does not contain "local-pg"
    And the keyring holds nothing for "local-pg"

  @R-01 @T-02-09
  Scenario: Switch connection from the shell and come back
    When I start vanitas with "-c notes"
    And the connection finishes
    And I press "space"
    Then the popup lists "c" with "switch connection"
    When I press "c"
    Then the screen shows "Connections"
    When I press "esc"
    Then the screen shows "vanitas ─ conn: notes ─ db: seed.db"

  @R-01 @T-02-09
  Scenario: Close the picker at startup and get help
    When I start vanitas without arguments
    And I press "?"
    Then the screen shows "Help — Connections"
    When I press "esc"
    And I press "esc"
    Then the screen shows "vanitas ─ no connection"

  @R-01 @R-LGPD @integration @T-02-09
  Scenario: A wrong password typed in the prompt returns to the picker
    Given the PostgreSQL 16 seed server
    And a saved connection "local-pg" to the PostgreSQL server as "app" without password
    When I start vanitas with "-c local-pg"
    And the connection finishes
    And I type "wrong-Pw-123"
    And I press "enter"
    And the connection finishes
    Then the screen shows "Connections"
    And the status bar contains "could not connect to local-pg:"
    And the screen does not contain "wrong-Pw-123"

  @R-01 @R-02 @integration @T-02-09
  Scenario Outline: Test and connect to each server from the picker
    Given the <server> seed server
    And a saved connection "srv" to the <server> server as "app" with password "app-Pw-123"
    When I start vanitas without arguments
    And I select "srv"
    And I press "t"
    And the test finishes
    Then the status bar contains "connection ok: <product>"
    When I press "enter"
    And the connection finishes
    Then the screen shows "vanitas ─ conn: srv ─ db: app"

    Examples:
      | server        | product    |
      | PostgreSQL 16 | PostgreSQL |
      | MySQL 8       | MySQL      |
      | MariaDB 11    | MariaDB    |

  @R-01 @T-02-10
  Scenario: Picker footer and layout
    When I start vanitas without arguments
    Then the footer is "enter connect  a add  e edit  d delete  t test  / filter  esc close  ? help"
    And the screen matches the golden file "connpicker_list"

  @R-01 @T-02-10
  Scenario: Add a SQLite connection
    When I start vanitas without arguments
    And I press "a"
    Then the screen shows "New connection"
    When I type "inbox"
    And I press "tab"
    And I press "right"
    And I press "right"
    And I press "tab"
    And I type the path of the SQLite file "seed.db"
    And I press "ctrl+s"
    And the save finishes
    Then the picker lists "notes", "local-pg", "shop" and "inbox" in this order
    And the status bar contains "saved inbox"
    And the connections file contains "driver = \"sqlite\""

  @R-01 @R-LGPD @T-02-10
  Scenario: Add a PostgreSQL connection with a password
    When I start vanitas without arguments
    And I press "a"
    And I fill the form with:
      | field    | value     |
      | Name     | team-pg   |
      | Host     | db.test   |
      | User     | app       |
      | Password | s3cr3t-Pw |
      | Database | app       |
    Then the screen shows "•••••••••"
    And the screen does not contain "s3cr3t-Pw"
    And the screen matches the golden file "connpicker_form"
    When I press "ctrl+s"
    And the save finishes
    Then the keyring holds "s3cr3t-Pw" for "team-pg"
    And the connections file contains "host = \"db.test\""
    And the connections file does not contain "s3cr3t-Pw"

  @R-01 @T-02-10
  Scenario Outline: Form validation
    When I start vanitas without arguments
    And I press "a"
    And I fill the form with:
      | field   | value   |
      | Name    | <name>  |
      | Port    | <port>  |
      | Color   | <color> |
    And I press "ctrl+s"
    Then the screen shows "<message>"
    And the screen shows "New connection"

    Examples:
      | name   | port  | color | message                                                                                          |
      |        | 5432  |       | name is required                                                                                 |
      | my db  | 5432  |       | name must start with a letter or digit and use only letters, digits, ".", "_" or "-" (max 64)   |
      | notes  | 5432  |       | a connection named "notes" already exists                                                       |
      | new-pg | 99999 |       | port must be a number from 1 to 65535                                                           |
      | new-pg | 5432  | red   | color must be #RRGGBB                                                                           |

  @R-01 @T-02-10
  Scenario: Edit keeps the stored password when the field is left empty
    When I start vanitas without arguments
    And I press "down"
    And I press "e"
    Then the screen shows "Edit connection"
    And the screen shows "stored in keyring"
    When I fill the form with:
      | field | value |
      | Port  | 5433  |
    And I press "ctrl+s"
    And the save finishes
    Then the connections file contains "port = 5433"
    And the keyring holds "s3cr3t-Pw" for "local-pg"

  @R-01 @T-02-10
  Scenario: Renaming moves the stored password
    When I start vanitas without arguments
    And I press "down"
    And I press "e"
    And I fill the form with:
      | field | value   |
      | Name  | main-pg |
    And I press "ctrl+s"
    And the save finishes
    Then the keyring holds "s3cr3t-Pw" for "main-pg"
    And the keyring holds nothing for "local-pg"

  @R-01 @R-LGPD @T-02-10
  Scenario: Saving without a keyring warns and keeps no password
    Given the keyring is unavailable
    When I start vanitas without arguments
    And I press "a"
    And I fill the form with:
      | field    | value     |
      | Name     | team-pg   |
      | Password | s3cr3t-Pw |
    And I press "ctrl+s"
    And the save finishes
    Then the status bar contains "saved team-pg; keyring unavailable, the password will be asked when connecting"
    And the connections file does not contain "s3cr3t-Pw"

  @R-01 @R-LGPD @T-02-10
  Scenario: Templates are edited as written
    Given the saved connection "prod" with user "${PROD_USER}" and host "10.0.0.5" for driver "mysql"
    When I start vanitas without arguments
    And I select "prod"
    And I press "e"
    Then the screen shows "${PROD_USER}"
    When I press "ctrl+s"
    And the save finishes
    Then the connections file contains "user = \"${PROD_USER}\""

  @R-01 @T-02-10
  Scenario: Switching the driver updates an untouched port
    When I start vanitas without arguments
    And I press "a"
    And I press "tab"
    Then the form field "Port" is "5432"
    When I press "right"
    Then the form field "Port" is "3306"

  @R-01 @T-02-10
  Scenario: Test and cancel from the form
    When I start vanitas without arguments
    And I press "e"
    And I press "ctrl+t"
    And the test finishes
    Then the status bar contains "connection ok: SQLite 3."
    And the screen shows "Edit connection"
    When I press "esc"
    Then the screen shows "Connections"
    And the screen does not show "Edit connection"
