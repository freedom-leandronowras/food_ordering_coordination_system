Feature: Place order success
  As a member
  I want to place an order with enough credits
  So that my order is confirmed and my credits are deducted

  Scenario: Member places a valid order with enough credits
    Given a member has 100 credits
    When the member places an order totaling 30
    Then the order is confirmed
    And remaining credits should be 70
    And an order-created event exists
