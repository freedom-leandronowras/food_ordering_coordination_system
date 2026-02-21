# Overview
For a office setting:
This is a system for coordinating food orders to solve the problem of:

Different companies have their innovation departments located in the same physical place, and there are two HIVE managers who make sure everything runs smoothly.
Currently, there is no  -> coordination <- of food ordering.

should it be a system to make orders or to setup another system that make orders ( i will go with the first one)

- Features
    - order food


## Assumptions
- other systems already have their database
    - so i wont touch it for now

- the data of the external system is accurated

- if an external system replies with 200 code, the order is confirmed

### Events

### Domain
- 



## Approach
- [ ] is it a event driven architecture?

- It focus on coordenation, not external services
- focus on persist events instead of domain
- Should have a layer for external integrations

- Build focusing on allowing integration with new services without breaking new ones
- in memory mock db to:
    -  how do i know the items?
    -  How do i know the stock

- Build for adaptability (add new services)
- Focus on events not domain 
- focus on build fast for fast feedback loop

### Design/Architecture
- Queue or Lock? (which one is easier in go?)
    - use enterprise integration patterns the basics
        - but keep it simple and with graphql first
- RBCA

- persist events not domain entities
- Adapter pattern
- Graphql for different resolvers (companies/services)

### Frameworks
#### Backend
- Go
    - because its fast iteration and concurrency features
- Serveless functions
- GraphQl
- validation framework
- Clerk SSO for Auth/Authentication

#### WebUi
- Nextjs

#### Persistance
- Dash0

### Deployment

### trade-offs


---
# What you’d do next with more time
- queue instead of lock?
- resilience queue and microservices patterns

- email notification

- strategy for each company and role
    - and food of the day, cupons, etc
- group social activities
- AI integration with mira vision



---
# Future Functionalities
What potential functionalities or features might be useful in the future?

Please describe:
- Which features could be added.
- How they could be integrated into your existing implementation.
- Whether architectural changes would be necessary to support them.


- uber eats integration

---
# Notion
- Overview of your approach
- Assumptions you made
- Architecture / diagrams / sketches (if helpful)
- Any code snippets or links (GitHub optional, not required)
- Decisions & trade-offs
- What you’d do next with more time
