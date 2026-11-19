Feature: Customisable keymap
  Users override default shortcuts per scope in keymap.toml.

  @R-12 @T-01-04
  Scenario: Default bindings without a keymap file
    Given no keymap file
    When the keymap is loaded
    Then there are no keymap warnings
    And in scope "global" the key "ctrl+q" runs "quit"
    And in scope "global" the key "?" runs "help"
    And in scope "leader" the key "t" runs "switch_theme"

  @R-12 @T-01-04
  Scenario: A user override replaces the default key
    Given a keymap file containing:
      """
      [global]
      help = "f1"
      """
    When the keymap is loaded
    Then there are no keymap warnings
    And in scope "global" the key "f1" runs "help"
    And in scope "global" the key "?" runs nothing

  @R-12 @T-01-04
  Scenario: An action can have several keys
    Given a keymap file containing:
      """
      [global]
      quit = ["ctrl+q", "f10"]
      """
    When the keymap is loaded
    Then in scope "global" the key "ctrl+q" runs "quit"
    And in scope "global" the key "f10" runs "quit"

  @R-12 @T-01-04
  Scenario: "none" removes a binding
    Given a keymap file containing:
      """
      [leader]
      switch_theme = "none"
      """
    When the keymap is loaded
    Then there are no keymap warnings
    And in scope "leader" the key "t" runs nothing
    And scope "leader" does not list the action "switch_theme"

  @R-12 @T-01-04
  Scenario: A conflict in the same scope is rejected with a warning
    Given a keymap file containing:
      """
      [global]
      help = "ctrl+q"
      """
    When the keymap is loaded
    Then there is the keymap warning "keymap: [global] key \"ctrl+q\" for \"help\" conflicts with \"quit\"; override ignored"
    And in scope "global" the key "ctrl+q" runs "quit"
    And in scope "global" the key "?" runs "help"

  @R-12 @T-01-04
  Scenario: Two overridden actions can swap keys
    Given a keymap file containing:
      """
      [global]
      next_panel = "shift+tab"
      prev_panel = "tab"
      """
    When the keymap is loaded
    Then there are no keymap warnings
    And in scope "global" the key "tab" runs "prev_panel"
    And in scope "global" the key "shift+tab" runs "next_panel"

  @R-12 @T-01-04
  Scenario: The same key in different scopes is not a conflict
    Given a keymap file containing:
      """
      [leader]
      quit = "x"
      [help]
      close = "x"
      """
    When the keymap is loaded
    Then there are no keymap warnings
    And in scope "leader" the key "x" runs "quit"
    And in scope "help" the key "x" runs "close"

  @R-12 @T-01-04
  Scenario: Unknown scopes and actions are reported
    Given a keymap file containing:
      """
      [sidebar]
      open = "o"
      [global]
      launch_rockets = "ctrl+l"
      """
    When the keymap is loaded
    Then there is the keymap warning "keymap: unknown scope \"sidebar\""
    And there is the keymap warning "keymap: unknown action \"launch_rockets\" in [global]"

  @R-12 @T-01-04
  Scenario: Invalid TOML in the keymap file is an error
    Given a keymap file containing:
      """
      [global
      help = "f1"
      """
    When the keymap is loaded
    Then loading the keymap fails with an error containing "keymap.toml"
