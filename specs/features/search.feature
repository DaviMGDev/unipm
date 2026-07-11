# language: en

Feature: Search across package sources
  As a sysadmin or developer
  I want to search for a package by name across all available backends in parallel
  So that I can find where a package exists without running separate search commands

  Background:
    Given unipm is installed and at least two backends are available

  @smoke @search
  Scenario: Basic search returns results from all backends
    Given the "apt" backend has "htop"
    And the "brew" backend has "htop"
    When the user runs "unipm search htop"
    Then the output SHALL contain a table with columns Source, Name, Version, Description
    And the table SHALL include a row for "apt" with "htop"
    And the table SHALL include a row for "brew" with "htop"
    And the results SHALL be deduplicated by "(Source, Name)"

  @search
  Scenario: Search finds package in exactly one backend
    Given the "apt" backend has "ripgrep"
    And no other backend has "ripgrep"
    When the user runs "unipm search ripgrep"
    Then the output SHALL contain exactly one row for "apt" with "ripgrep"

  @search @error
  Scenario: No backends are available
    Given no package manager binaries are on $PATH
    When the user runs "unipm search htop"
    Then the output SHALL contain an error message listing unavailable adapters
    And the exit code SHALL be non-zero

  @search @timeout
  Scenario: One backend times out while others succeed
    Given the "apt" backend is available
    And the "npm" backend is unresponsive
    When the user runs "unipm search htop"
    Then the output SHALL include results from "apt"
    And the output SHALL include a warning for "npm" timeout
    And the exit code SHALL be zero

  @search
  Scenario Outline: Search with custom timeout
    Given the "apt" backend has "<pkg>"
    When the user runs "unipm search <pkg> --timeout <timeout>"
    Then each backend SHALL be given <timeout> seconds to respond

    Examples:
      | pkg      | timeout |
      | htop     | 5       |
      | ripgrep  | 15      |
