# System Architecture: Food Ordering Coordination System

This document outlines the high-level design, technical patterns, and architectural decisions governing the Food Ordering Coordination System.

---

## 📁 Project Structure & Decoupling

The codebase is organized into three distinct top-level directories to ensure a clean separation of concerns and facilitate independent scaling and development.

### 1. `web_ui/` (Frontend)
*   **Purpose:** A modern **Next.js** application that serves as the primary user interface.
*   **Rationale:** By isolating the UI, we can iterate on the user experience, styling, and client-side logic without redeploying the backend. It communicates with the API via standard HTTP/JSON, treating it as a black box.

### 2. `coordination_api/` (Core Backend)
*   **Purpose:** The central **Go** service responsible for business logic, persistence, and vendor aggregation.
*   **Rationale:** Centralizing the "brain" of the system in a typed, high-performance language allows for robust transaction management and concurrent fan-out operations. This service maintains the source of truth for credits and orders.

### 3. `external_services_mocks/` (Infrastructure Simulation)
*   **Purpose:** Contains static JSON data and configurations used to simulate external vendor APIs (via `json-server`).
*   **Rationale:** This decoupling allows for **Offline Development**. Developers can test the Aggregator and Adapter patterns without requiring access to live vendor systems, ensuring the coordination logic is resilient to various external responses.

---

## 🏗️ Design/Architecture

### 1. High-Level System Flow
1.  **Identity:** The **Web UI** authenticates the user via **Clerk**. The issued JWT contains the `member_id` (sub) and `role`.
2.  **Discovery:** When a member opens the app, the UI calls `/api/menus`. The **Aggregator** triggers a **Fan-out** to all configured vendor **Adapters** concurrently.
3.  **Aggregation:** The results are **Fanned-in**, normalized into a single schema, and returned to the UI.
4.  **Transaction:** When an order is placed, the **Domain Service** performs a "Triple Write": it updates the current **Member Credits**, saves the **Food Order**, and appends an immutable **Event** to the audit trail.
5.  **Integration:** While the **Aggregator** supports concurrent order submission to external vendor APIs, the current implementation focuses on internal state management and auditability.

---

### 2. Core Architectural Pillars

#### Web UI (Next.js & React 19)
*   **Reactive State Management:** Uses a centralized `MenuProvider` (React Context) to coordinate the "Tray" logic, real-time credit deductions, and modal states.
*   **Optimistic Session Validation:** Performs client-side JWT parsing to provide immediate UI feedback and role-based navigation before the request even reaches the backend.
*   **Modular Component Architecture:** Separates the **App Shell** (layout) from **Domain Sections** (Tray, Menu, Dialogs) for high maintainability.

#### Fan-out/Fan-in Aggregator
*   **Concurrent Execution:** Utilizes Go **Goroutines** to request menus from multiple vendors in parallel, preventing a "waterfall" delay where one slow vendor blocks the others.
*   **Synchronization:** Employs `sync.WaitGroup` and Go **Channels** for fan-in, ensuring the API response is only sent once all vendors have responded or timed out.
*   **Fault Tolerance:** Implements per-vendor timeouts (10s); if one vendor fails, the aggregator returns a partial success with the remaining available menus.

#### Adapter Pattern
*   **External Decoupling:** Defines a standard interface for `FetchMenu` and `SubmitOrder`, ensuring the core business logic never knows if it's talking to a JSON server, a legacy SQL DB, or a 3rd party API.
*   **Data Translation:** The `JSONServerAdapter` acts as a translator, converting external "Wire Types" (raw JSON) into internal **Domain Entities** to keep the core code clean and typed.

#### Hybrid Persistence
*   **Performance Reads:** Stores the latest **Snapshots** for `credits` and `orders` in dedicated MongoDB collections. This allows the UI to fetch history and balances with sub-millisecond latency.
*   **Immutable Event Store:** Every mutation (credit adjustments, order placement) appends a record to the `events` collection. This provides a "Time Machine" for the system, allowing for perfect auditing.
*   **Future-Proofing:** By storing events today, the system is ready to migrate to **Full Event Sourcing** or trigger **Side-effects** (like email notifications) without changing the core repository.

#### RBAC (Role-Based Access Control)
*   **Middleware Enforcement:** A specialized `Authenticator` higher-order function wraps every API route, checking the `role` claim in the JWT before allowing the request to hit the controller.
*   **Granular Permissions:**
    *   `MEMBER`: Can view menus and place orders.
    *   `HIVE_MANAGER`: Can view all data and **Grant Credits** via protected POST endpoints.
*   **Identity Validation:** The system validates that the `member_id` in the request body matches the secure JWT `sub` claim for non-managerial requests, ensuring a user can only perform actions on their own behalf.

---

### 3. Event Catalog & Data Integrity

The system treats **Events** as the source of truth for all state mutations.

*   **`food-order.created.v1`**
    *   **Trigger:** When `PlaceOrderUseCase` successfully validates credits and saves a new order.
    *   **Payload:** A complete snapshot of the `FoodOrder` including `OrderID`, `MemberID`, itemized list with prices, and `TotalPrice`.
*   **`credits.granted.v1`**
    *   **Trigger:** When a manager (`HIVE_MANAGER`) manually adjusts a member's balance.
    *   **Payload:** Captures the `MemberID`, the `Amount` granted, and the resulting `NewBalance`.
*   **Causal Metadata**
    *   The event schema includes `CorrelationID` and `CausationID` to support future tracing of complex user-driven side-effects.

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
