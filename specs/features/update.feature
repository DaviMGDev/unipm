# language: en

Feature: Update tracked packages
  As a sysadmin
  I want to update packages I installed via unipm, either individually or all at once
  So that my system stays current without checking each backend manually

  Background:
    Given unipm is installed
    And state.json contains three entries: htop (apt), httpie (pypi), rg (brew)

  @smoke @update
  Scenario: Update all tracked packages
    When the user runs "unipm update"
    Then the system SHALL read all entries from state.json
    And the system SHALL delegate to "sudo apt upgrade htop" for the apt entry
    And the system SHALL delegate to "pip3 install --upgrade httpie" for the pypi entry
    And the system SHALL delegate to "brew upgrade rg" for the brew entry
    And the system SHALL refresh the version field for each successfully updated package

  @update
  Scenario: Update a single tracked package
    When the user runs "unipm update htop"
    Then the system SHALL look up "htop" in state.json
    And the system SHALL delegate only to the "apt" backend for the update
    And the system SHALL refresh the version field for "htop"

  @update @error
  Scenario: Update a package not tracked by unipm
    Given state.json does NOT contain an entry for "neovim"
    When the user runs "unipm update neovim"
    Then the output SHALL contain: "neovim was not installed via unipm"
    And the exit code SHALL be non-zero

  @update @error
  Scenario: One package fails during "update all" — others continue
    Given the "apt" backend update for "htop" fails with a network error
    When the user runs "unipm update"
    Then the system SHALL report: "✗ htop (apt): <error message>"
    And the system SHALL continue updating "httpie" and "rg"
    And the exit code SHALL be non-zero (partial failure)
    And successfully updated packages SHALL have their version field refreshed
