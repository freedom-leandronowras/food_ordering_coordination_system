# System Architecture: Food Ordering Coordination System

This document outlines the high-level design, technical patterns, and architectural decisions governing the Food Ordering Coordination System.

---

## 🏗️ Design/Architecture

### 1. High-Level System Flow
1.  **Identity:** The **Web UI** authenticates the user via **Clerk**. The issued JWT contains the `member_id` (sub) and `role`.
2.  **Discovery:** When a member opens the app, the UI calls `/api/menus`. The **Aggregator** triggers a **Fan-out** to all configured vendor **Adapters** concurrently.
3.  **Aggregation:** The results are **Fanned-in**, normalized into a single schema, and returned to the UI.
4.  **Transaction:** When an order is placed, the **Domain Service** performs a "Double Write": it updates the current **State** (credits/orders) for performance and appends an immutable **Event** for the audit trail.
5.  **Integration:** The **Adapter** finally submits the order to the external vendor's system to close the loop.

---

### 2. Core Architectural Pillars

#### Web UI (Next.js & React 19)
*   **Reactive State Management:** Uses a centralized `MenuProvider` (React Context) to coordinate the "Tray" logic, real-time credit deductions, and modal states.
*   **Optimistic Session Validation:** Performs client-side JWT parsing to provide immediate UI feedback and role-based navigation before the request even reaches the backend.
*   **Modular Component Architecture:** Separates the **App Shell** (layout) from **Domain Sections** (Tray, Menu, Dialogs) for high maintainability.

#### Fan-out/Fan-in Aggregator
*   **Concurrent Execution:** Utilizes Go **Goroutines** to request menus from multiple vendors in parallel, preventing a "waterfall" delay where one slow vendor blocks the others.
*   **Synchronization:** Employs `sync.WaitGroup` and thread-safe slices to collect results, ensuring the API response is only sent once all vendors have responded or timed out.
*   **Fault Tolerance:** Implements per-vendor timeouts (10s); if one vendor fails, the aggregator returns a partial success with the remaining available menus.

#### Adapter Pattern
*   **External Decoupling:** Defines a standard interface for `FetchMenu` and `SubmitOrder`, ensuring the core business logic never knows if it's talking to a JSON server, a legacy SQL DB, or a 3rd party API.
*   **Data Translation:** The `JSONServerAdapter` acts as a translator, converting external "Wire Types" (raw JSON) into internal **Domain Entities** to keep the core code clean and typed.

#### Hybrid Persistence
*   **Performance Reads:** Stores the latest "Snapshot" of credits and orders in dedicated MongoDB collections. This allows the UI to fetch history and balances with sub-millisecond latency.
*   **Immutable Event Store:** Every mutation appends a record to the `events` collection. This provides a "Time Machine" for the system, allowing for perfect auditing of credit grants and order creation.
*   **Future-Proofing:** By storing events today, the system is ready to migrate to **Full Event Sourcing** or trigger **Side-effects** (like email notifications) without changing the core repository.

#### RBAC (Role-Based Access Control)
*   **Middleware Enforcement:** A specialized `Authenticator` higher-order function wraps every API route, checking the `role` claim in the JWT before allowing the request to hit the controller.
*   **Granular Permissions:**
    *   `MEMBER`: Can view menus and place orders.
    *   `HIVE_MANAGER`: Can view all data and **Grant Credits** via protected POST endpoints.
*   **Zero-Trust Identity:** The system derives the `member_id` directly from the secure JWT `sub` claim rather than trusting a user-provided ID in the request body.

---

### 3. Event Catalog & Data Integrity

The system treats **Events** as the source of truth for all state mutations.

*   **`food-order.created.v1`**
    *   **Trigger:** When `PlaceOrderUseCase` successfully validates credits and saves a new order.
    *   **Payload:** A complete snapshot of the `FoodOrder` including `OrderID`, `MemberID`, itemized list with prices, and `TotalPrice`.
*   **`credits.granted.v1`**
    *   **Trigger:** When a `HIVE_MANAGER` manually adjusts a member's balance.
    *   **Payload:** Captures the `MemberID`, the `Amount` granted, and the resulting `NewBalance`.
*   **Causal Metadata**
    *   Every event includes a `CorrelationID` and `CausationID` to trace a single user action through every side-effect it triggered.

---

### 4. Domain Logic & Use Case Pattern

The core logic is isolated using the **Clean Architecture** use-case pattern.

*   **Business Rules over Data:** The `domain` package contains no knowledge of MongoDB or JSON. It only knows about entities like `Member`, `Credit`, and `FoodOrder`.
*   **Transactional Integrity:** The `PlaceOrderUseCase` ensures an atomic transaction logic: credit check, deduction, order save, and event append—all or nothing.
*   **Domain Constraints:** Enforces a strict `MaxMemberCredits` (1,000) at the domain level to prevent accidental or malicious over-funding.

---

### 5. Technical Trade-offs & Rationale

*   **Hybrid Persistence vs. Pure Event Sourcing**
    *   *Trade-off:* Storing "Current State" alongside "Events."
    *   *Rationale:* Pure Event Sourcing is complex to query. Snapshotting state in a `credits` collection provides **O(1) read performance** for the UI while keeping **100% Auditability**.
*   **Synchronous Vendor Aggregation**
    *   *Trade-off:* The frontend waits for the Go API, which waits for all external vendors.
    *   *Rationale:* To provide a "Single Pane of Glass" experience. We mitigate delays using **Parallel Goroutines** and a **10-second timeout**.
*   **Deferred JWT Verification**
    *   *Trade-off:* The API currently trusts JWT claims after parsing without signature validation.
    *   *Rationale:* Accelerates the feedback loop during prototyping. The architecture allows swapping the `Authenticator` for a full-verification version without touching business logic.
*   **In-Memory vs. Persistent Mocking**
    *   *Trade-off:* Using `json-server` with flat files for vendors instead of a real database.
    *   *Rationale:* Makes the ecosystem **Portable**. Developers can clone and run a multi-vendor environment without complex infrastructure setup.
