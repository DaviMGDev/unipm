# language: en

Feature: Uninstall a tracked package
  As a sysadmin or developer
  I want to uninstall a package via unipm and have it automatically route to the correct backend
  So that I don't need to remember which package manager I originally used

  Background:
    Given unipm is installed
    And "~/.unipm/state.json" exists

  @smoke @uninstall
  Scenario: Uninstall a package tracked in state.json
    Given state.json contains: { "name": "htop", "source": "apt", "version": "3.3.0" }
    When the user runs "unipm uninstall htop"
    Then the system SHALL delegate to "sudo apt remove htop"
    And the system SHALL remove the "htop" record from state.json
    And the system SHALL print "✓ htop removed from apt"

  @uninstall @error
  Scenario: Uninstall a package not tracked by unipm
    Given state.json does NOT contain an entry for "htop"
    When the user runs "unipm uninstall htop"
    Then the output SHALL contain: "htop was not installed via unipm"
    And the exit code SHALL be non-zero

  @uninstall @error
  Scenario: Backend removal fails
    Given state.json contains: { "name": "htop", "source": "apt", "version": "3.3.0" }
    And "sudo apt remove htop" exits with an error
    When the user runs "unipm uninstall htop"
    Then the output SHALL display the native error from apt
    And the output SHALL offer: "Remove htop from unipm tracking anyway? [y/N]"
    And the package record SHALL remain in state.json unless user confirms removal

  @uninstall
  Scenario: Uninstall a package installed from pypi (user scope)
    Given state.json contains: { "name": "httpie", "source": "pypi", "version": "3.2.1" }
    When the user runs "unipm uninstall httpie"
    Then the system SHALL delegate to "pip3 uninstall httpie"
    And the system SHALL remove the "httpie" record from state.json
