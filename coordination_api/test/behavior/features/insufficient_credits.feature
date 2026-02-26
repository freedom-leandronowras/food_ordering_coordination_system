Feature: Insufficient credits
  As a member
  I should be blocked when I cannot afford an order
  So that credits never become negative

  Scenario: Member tries to place an order without enough credits
    Given a member has 10 credits
    When the member places an order totaling 25
    Then the request is rejected with "INSUFFICIENT_CREDITS"
    And credits remain 10
    And no order or event is created
