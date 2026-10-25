Feature: Configuration directory and settings
  Vanitas keeps its files in one configuration directory and reads settings.toml.

  @R-16 @T-01-03
  Scenario: VANITAS_CONFIG_DIR overrides every other location
    Given the environment variable "VANITAS_CONFIG_DIR" is "/tmp/vt-conf"
    And the environment variable "XDG_CONFIG_HOME" is "/tmp/xdg"
    When the configuration paths are resolved
    Then the configuration directory is "/tmp/vt-conf"
    And the settings file is "/tmp/vt-conf/settings.toml"
    And the keymap file is "/tmp/vt-conf/keymap.toml"

  @R-16 @T-01-03
  Scenario: XDG_CONFIG_HOME is used when VANITAS_CONFIG_DIR is not set
    Given the environment variable "VANITAS_CONFIG_DIR" is ""
    And the environment variable "XDG_CONFIG_HOME" is "/tmp/xdg"
    When the configuration paths are resolved
    Then the configuration directory is "/tmp/xdg/vanitas"

  @R-16 @T-01-03
  Scenario: Missing settings file gives the defaults
    Given an empty configuration directory
    When the settings are loaded
    Then loading succeeds with no warnings
    And the setting "theme" is "vanitas"
    And the setting "page_size" is "500"
    And the setting "query_timeout" is "0s"
    And the setting "vim_mode" is "true"
    And no file was created in the configuration directory

  @R-16 @T-01-03
  Scenario: Values from settings.toml replace the defaults
    Given a settings file containing:
      """
      theme = "nord"
      page_size = 200
      query_timeout = "30s"
      zebra = false
      """
    When the settings are loaded
    Then loading succeeds with no warnings
    And the setting "theme" is "nord"
    And the setting "page_size" is "200"
    And the setting "query_timeout" is "30s"
    And the setting "zebra" is "false"
    And the setting "fk_display" is "true"

  @R-16 @T-01-03
  Scenario: Unknown keys are reported as warnings
    Given a settings file containing:
      """
      theme = "nord"
      colour_scheme = "blue"
      """
    When the settings are loaded
    Then loading succeeds with the warning "settings.toml: unknown key \"colour_scheme\""
    And the setting "theme" is "nord"

  @R-16 @T-01-03
  Scenario: Invalid TOML is an error naming the file and line
    Given a settings file containing:
      """
      theme = "nord"
      page_size = = 3
      """
    When the settings are loaded
    Then loading fails with an error containing "settings.toml"
    And the error contains "line 2"

  @R-16 @T-01-03
  Scenario: Out-of-range page size is an error
    Given a settings file containing:
      """
      page_size = 0
      """
    When the settings are loaded
    Then loading fails with an error containing "page_size must be between 1 and 100000, got 0"

  @R-16 @T-01-03
  Scenario: Saving settings writes the file atomically
    Given an empty configuration directory
    When the settings are saved with "theme" set to "dracula"
    And the settings are loaded
    Then the setting "theme" is "dracula"
    And the configuration directory contains only "settings.toml"
    And the file "settings.toml" has permissions "0600"
