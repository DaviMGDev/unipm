# language: en

Feature: Install a package with collision resolution
  As a sysadmin or developer
  I want to install a package via unipm and be prompted to choose a source when ambiguous
  So that I never accidentally install from the wrong ecosystem

  Background:
    Given unipm is installed
    And the "apt" and "pypi" backends are available

  @smoke @install
  Scenario: Install when exactly one backend has the package
    Given only the "apt" backend has "htop"
    When the user runs "unipm install htop"
    Then the system SHALL delegate to "sudo apt install htop" without prompting
    And the system SHALL record the package in "~/.unipm/state.json" with source "apt", name "htop"
    And the system SHALL print "✓ htop <version> installed from apt"

  @install @tui
  Scenario: Install when multiple backends have the package — TUI collision resolution
    Given the "apt" backend has "requests"
    And the "pypi" backend has "requests"
    When the user runs "unipm install requests"
    Then the system SHALL open an interactive TUI listing:
      | Source | Name    | Version | Description                |
      | apt    | requests | ...    | ...                        |
      | pypi   | requests | ...    | Python HTTP for Humans     |
    And the sources SHALL be listed alphabetically by source name
    And no source SHALL be preselected
    When the user selects "pypi" and presses Enter
    Then the system SHALL delegate to "pip3 install --user requests"
    And the system SHALL record the package in "~/.unipm/state.json" with source "pypi"

  @install
  Scenario: Install directly from a named source with --source flag
    Given the "apt" backend has "htop"
    When the user runs "unipm install htop --source apt"
    Then the system SHALL install from "apt" without any prompt or search
    And the system SHALL validate that "apt" is an available adapter

  @install
  Scenario: Install from multiple sources with --source flag
    Given the "apt" backend has "htop"
    And the "brew" backend has "htop"
    When the user runs "unipm install htop --source apt,brew"
    Then the system SHALL install from "apt"
    And the system SHALL install from "brew"
    And the system SHALL record both installations in state.json

  @install @error
  Scenario: No backend has the package
    Given no backend has "nonexistent-pkg-xyz"
    When the user runs "unipm install nonexistent-pkg-xyz"
    Then the output SHALL contain an error message suggesting a different query or "unipm sources"
    And the exit code SHALL be non-zero

  @install @error
  Scenario: Named source via --source is unavailable
    Given the "cargo" adapter is not available on this system
    When the user runs "unipm install ripgrep --source cargo"
    Then the output SHALL contain an error listing available sources
    And the exit code SHALL be non-zero

  @install @error
  Scenario: Native installation fails
    Given the "apt" backend has "htop"
    And "sudo apt install htop" exits with an error
    When the user runs "unipm install htop"
    Then the output SHALL display the native error message from apt
    And the package SHALL NOT be recorded in state.json
    And the exit code SHALL be non-zero
