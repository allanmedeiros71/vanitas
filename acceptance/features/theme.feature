Feature: Themes
  Vanitas ships eight themes made of named tokens, which users can override.

  @R-12 @T-01-05
  Scenario: The eight built-in themes in cycle order
    When I list the themes
    Then the themes are "vanitas, light, tokyo-night, rose-pine, gruvbox, nord, catppuccin-mocha, dracula"

  @R-12 @T-01-05
  Scenario Outline: Every theme defines every token
    When I load the theme "<theme>"
    Then every token has a "#RRGGBB" colour
    And the contrast of "fg" on "bg" is at least 4.5
    And the contrast of "focus" on "bg" is at least 3.0

    Examples:
      | theme            |
      | vanitas          |
      | light            |
      | tokyo-night      |
      | rose-pine        |
      | gruvbox          |
      | nord             |
      | catppuccin-mocha |
      | dracula          |

  @R-12 @T-01-05
  Scenario: Unknown theme name
    When I load the theme "solarized"
    Then loading the theme fails with the error "unknown theme \"solarized\""

  @R-12 @T-01-05
  Scenario: A token override replaces the theme colour
    When I load the theme "nord" with the override "zebra" = "#112233"
    Then the token "zebra" is "#112233"
    And the token "fg" is the same as in the theme "nord"

  @R-12 @T-01-05
  Scenario: Invalid overrides are reported and ignored
    When I load the theme "nord" with the override "zebra" = "blue"
    Then there is the theme warning "theme_overrides: invalid colour \"blue\" for \"zebra\""
    And the token "zebra" is the same as in the theme "nord"

  @R-12 @T-01-05
  Scenario: Unknown override token is reported
    When I load the theme "nord" with the override "sparkle" = "#112233"
    Then there is the theme warning "theme_overrides: unknown token \"sparkle\""

  @R-12 @T-01-05
  Scenario: NO_COLOR removes every colour
    Given NO_COLOR is set
    When I render a focused panel titled "Explorer" with the theme "dracula"
    Then the output has no colour escape sequences
    And the panel border is thick
