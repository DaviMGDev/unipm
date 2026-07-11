# language: en

Feature: Shell tab-completion
  As a developer
  I want tab-completion for commands, source names, and package names
  So that I can use unipm as fast as I type

  Background:
    Given unipm is installed
    And the shell completion script has been sourced

  @smoke @completion
  Scenario: Generate completion script for bash
    When the user runs "unipm completion bash"
    Then the output SHALL be a valid bash completion script
    And the output SHALL complete subcommands: search, install, uninstall, update, sources, completion

  @completion
  Scenario Outline: Generate completion script for each supported shell
    When the user runs "unipm completion <shell>"
    Then the output SHALL be a valid completion script for <shell>

    Examples:
      | shell |
      | bash  |
      | zsh   |
      | fish  |

  @completion
  Scenario: Tab-complete --source flag with available adapters
    Given the "apt" and "npm" adapters are available
    When the user types "unipm install htop --source " and presses TAB
    Then the completion SHALL suggest "apt" and "npm"
    And the completion SHALL NOT suggest unavailable adapters (e.g., "brew" if not on PATH)

  @completion
  Scenario: Tab-complete package names from local cache
    Given "~/.unipm/cache.json" contains package names: ["htop", "httpie", "ripgrep"]
    And the cache TTL has not expired
    When the user types "unipm install ht" and presses TAB
    Then the completion SHALL suggest "htop" and "httpie"
    And the completion SHALL NOT suggest "ripgrep"

  @completion
  Scenario: No network completions for queries shorter than 3 characters
    Given the local cache is empty or expired
    When the user types "unipm install ht" and presses TAB
    Then the system SHALL NOT make network requests for completions
    And no completions SHALL be returned (or only local cache matches)

  @completion
  Scenario: Completions for --source flag filter by partial input
    Given the "apt", "npm", and "appimage" adapters are available
    When the user types "unipm install htop --source ap" and presses TAB
    Then the completion SHALL suggest "apt" and "appimage"
    And the completion SHALL NOT suggest "npm"
