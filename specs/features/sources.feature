# language: en

Feature: List available package sources
  As a sysadmin or developer
  I want to see which package sources are detected and available on my system
  So that I know which backends unipm can use before I search or install

  Background:
    Given unipm is installed

  @smoke @sources
  Scenario: List all available and unavailable adapters
    Given "apt" is on $PATH
    And "npm" is on $PATH
    And "brew" is NOT on $PATH
    And "flatpak" is on $PATH
    When the user runs "unipm sources"
    Then the output SHALL display:
      """
      apt       ✓ available
      npm       ✓ available
      pypi      ✓ available (if pip3 on PATH)
      flatpak   ✓ available
      brew      ✗ not found on $PATH
      """
    And each line SHALL show the adapter name left-aligned followed by its status

  @sources
  Scenario: Distrobox adapters are listed per configured container
    Given "~/.unipm/config.yaml" defines a distrobox container "arch-dev" with package manager "yay"
    And the "arch-dev" container exists
    When the user runs "unipm sources"
    Then the output SHALL include:
      """
      distrobox-arch-dev  ✓ available (yay)
      """

  @sources @error
  Scenario: No adapters are available at all
    Given no supported package manager binaries are on $PATH
    And no distrobox containers are configured
    When the user runs "unipm sources"
    Then the output SHALL suggest installing at least one supported package manager
    And the exit code SHALL be non-zero
