Feature: App shell
  Running vanitas shows the workspace with placeholder panels, focus, help, leader menu and quit.

  Background:
    Given an empty configuration directory
    And a terminal of 120x40

  @R-12 @T-01-07
  Scenario: Startup layout
    When vanitas starts
    Then the screen shows "vanitas ─ no connection"
    And the screen shows the panels "Explorer", "Editor" and "Results"
    And the screen shows "No tabs open"
    And the focused panel is "Explorer"
    And the footer contains "? help"
    And the footer contains "␣ menu"
    And the screen matches the golden file "shell_startup"

  @R-12 @T-01-07
  Scenario: Focus panels with Alt+number
    When vanitas starts
    And I press "alt+3"
    Then the focused panel is "Results"
    When I press "alt+2"
    Then the focused panel is "Editor"
    When I press "alt+1"
    Then the focused panel is "Explorer"

  @R-12 @T-01-07
  Scenario: Tab and Shift+Tab cycle the panels
    When vanitas starts
    And I press "tab"
    Then the focused panel is "Editor"
    When I press "tab"
    Then the focused panel is "Results"
    When I press "tab"
    Then the focused panel is "Explorer"
    When I press "shift+tab"
    Then the focused panel is "Results"

  @R-12 @T-01-07
  Scenario: Help modal for the focused panel
    When vanitas starts
    And I press "alt+3"
    And I press "?"
    Then the screen shows "Help — Results"
    And the screen shows "Global"
    And the screen matches the golden file "shell_help"
    When I press "esc"
    Then the screen does not show "Help — Results"
    And the focused panel is "Results"

  @R-12 @T-01-07
  Scenario: Leader popup
    When vanitas starts
    And I press "space"
    Then the screen shows "Menu"
    And the screen shows "toggle explorer"
    And the screen matches the golden file "shell_leader"
    When I press "esc"
    Then the screen does not show "toggle explorer"

  @R-12 @T-01-07
  Scenario: Quit from the leader menu
    When vanitas starts
    And I press "space"
    And I press "q"
    Then vanitas exits

  @R-12 @T-01-07
  Scenario: Quit with Ctrl+Q
    When vanitas starts
    And I press "ctrl+q"
    Then vanitas exits

  @R-12 @T-01-07
  Scenario: Hide and show the explorer
    When vanitas starts
    And I press "space"
    And I press "e"
    Then the screen does not show "Explorer"
    And the focused panel is "Editor"
    When I press "alt+1"
    Then the screen shows "Explorer"
    And the focused panel is "Explorer"

  @R-12 @T-01-07
  Scenario: Switch theme applies at once and persists
    When vanitas starts
    And I press "space"
    And I press "t"
    Then the status bar shows "theme: light"
    And the settings file has "theme" set to "light"

  @R-12 @T-01-07
  Scenario: Unknown theme in settings falls back to vanitas
    Given the settings file contains 'theme = "solarized"'
    When vanitas starts
    Then the status bar shows "unknown theme \"solarized\", using \"vanitas\""

  @R-12 @T-01-07
  Scenario: Keymap conflicts are shown at startup
    Given the keymap file contains "[global]" and 'help = "ctrl+q"'
    When vanitas starts
    Then the status bar shows "conflicts with \"quit\"; override ignored"
    And the footer contains "? help"

  @R-12 @T-01-07
  Scenario: A user keymap override reaches the footer and the app
    Given the keymap file contains "[global]" and 'help = "f1"'
    When vanitas starts
    Then the footer contains "f1 help"
    When I press "f1"
    Then the screen shows "Help — Explorer"

  @R-12 @T-01-07
  Scenario: Terminal smaller than 80x24
    When vanitas starts in a terminal of 70x20
    Then the screen shows "Terminal too small: 70x20 (minimum 80x24)"
    And the screen does not show "Explorer"
    When the terminal is resized to 120x40
    Then the screen shows "Explorer"

  @R-12 @T-01-07
  Scenario: Invalid settings file stops startup
    Given the settings file contains "page_size = = 3"
    When I run "vanitas"
    Then the exit code is 1
    And the error output contains "settings.toml"
    And the settings file still contains "page_size = = 3"

  @R-LGPD @T-01-07
  Scenario: No log file without VANITAS_LOG
    Given the environment variable "VANITAS_LOG" is ""
    When vanitas starts
    And I press "ctrl+q"
    Then the configuration directory has no "vanitas.log"
