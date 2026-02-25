# Overview

## Clerk Auth Setup
- Copy `.env.example` to `.env.local`.
- Set `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` and `CLERK_SECRET_KEY` with values from Clerk Dashboard.
- Run `pnpm dev` and access `/`. Unauthenticated users are redirected to `/auth`.

For a office setting:
This is a system for coordinating food orders to solve the problem of:

Different companies have their innovation departments located in the same physical place, and there are two HIVE managers who make sure everything runs smoothly.
Currently, there is no  -> coordination <- of food ordering.

- [ ] should it be a system to make orders or to setup another system that make orders ( i will go with the first one)

- Features
    - see the menu per vendor
        - [x] entity for menu and vendor
    - order food
    - credits
        - [x] botar alguma limite tipo 1000
    - RBAC
        - [x] entity for jwt
        - [x] entity for roles and permissions?
    - more than one item a time
    - managers can:
        - add more credits to members
        - create new food services
        - see dashboard 


---
- [ ] checar se o mongo ta persistindo os events de maneira correta


- [ ] use framework for jwt parse in the auth.go

- [ ] user can see only see the data that belongs to him
- [ ] managers can see data from all users


- [ ] revisar user stories
   - [ ]  add table to test excepetions in invalid order payload
- [ ] user stories para (ou eh melhor em unit tests?)
    - [ ] criar para user dashboard
    - [ ] criar para manager

- [ ] botar camadas para os adapters e coisas q fazer sentindo de acordo com assumptions

## Assumptions
- other systems already have their database
    - so i wont touch it for now

- the data of the external system is accurated
- external systems always have unique id of entities


- if an external system replies with 200 code, the order is confirmed

- system is always available when using this software
    - too much?

- no limit to order as long as its available?



### Events

### Domain
- 

## Approach
- every member has 100 credits per month

- time related Persistance are made in the event struct
    - createdAt time.Time

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
- payment system for credits

- queue instead of lock?
- resilience queue and microservices patterns

- email notification

- strategy for each company and role
    - and food of the day, cupons, etc
- group social activities
- AI integration with mira vision


- events for rbac and member managament
- Currency conversion

---
# Future Functionalities
What potential functionalities or features might be useful in the future?

Please describe:
- Which features could be added.
- How they could be integrated into your existing implementation.
- Whether architectural changes would be necessary to support them.


- uber eats integration

- separate between type of meal:
    - lunch, snacks, meal

- analysis for cost optimzation

- events for RBAC

- carrinho to save order?

- favorites 

- quick reorder

- dish of the day

---
# Notion
- Overview of your approach
- Assumptions you made
- Architecture / diagrams / sketches (if helpful)
- Any code snippets or links (GitHub optional, not required)
- Decisions & trade-offs
- What you’d do next with more time
