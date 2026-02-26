Feature: Invalid payload handling
  As an API consumer
  I should receive a validation error for malformed payloads
  So that invalid orders do not mutate state

  Scenario: Malformed order payload is rejected
    Given a malformed order payload
    When the malformed order is submitted
    Then the response status should be 400
    And state is unchanged
