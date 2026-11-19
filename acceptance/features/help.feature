Feature: Footer, help modal and leader popup
  Shortcuts come from the keymap registry, so every binding is always discoverable.

  @R-12 @T-01-06
  Scenario: Footer shows the context's shortcuts in order
    Given the bindings "alt+1 explorer", "alt+2 editor", "? help" and "space menu"
    When the footer is rendered 120 columns wide
    Then the footer is "alt+1 explorer  alt+2 editor  ? help  ␣ menu"

  @R-12 @T-01-06
  Scenario: Footer shows at most 8 shortcuts
    Given 10 bindings "f1 one" to "f10 ten"
    When the footer is rendered 200 columns wide
    Then the footer shows 8 shortcuts
    And the footer does not contain "f9"

  @R-12 @T-01-06
  Scenario: Footer drops shortcuts that do not fit
    Given the bindings "alt+1 explorer", "alt+2 editor", "? help" and "space menu"
    When the footer is rendered 30 columns wide
    Then the footer is "alt+1 explorer  alt+2 editor"

  @R-12 @T-01-06
  Scenario: Every default binding appears in its scope's help
    When I open the help for each scope of the default keymap
    Then every binding of that scope is listed with its key and description
    And every "global" binding is listed under "Global"

  @R-12 @T-01-06
  Scenario: Help filter narrows the list
    Given the help modal for scope "global"
    When I type "quit"
    Then the help lists "ctrl+q" with "quit"
    And the help does not list "help"

  @R-12 @T-01-06
  Scenario: Help filter without matches
    Given the help modal for scope "global"
    When I type "zzz"
    Then the help shows "No matching shortcuts"

  @R-12 @T-01-06
  Scenario: Esc closes the help modal
    Given the help modal for scope "global"
    When I press "esc"
    Then the help modal reports it closed

  @R-12 @T-01-06
  Scenario: Leader popup lists its commands
    Given the leader popup with the default keymap
    Then the popup title is "Menu"
    And the popup lists "e" with "toggle explorer"
    And the popup lists "t" with "switch theme"
    And the popup lists "q" with "quit"

  @R-12 @T-01-06
  Scenario: Leader popup runs the chosen command
    Given the leader popup with the default keymap
    When I press "t"
    Then the popup reports the action "switch_theme"

  @R-12 @T-01-06
  Scenario: Other keys close the leader popup without an action
    Given the leader popup with the default keymap
    When I press "z"
    Then the popup reports it closed without an action
