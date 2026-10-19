# Aupa Dispatch Worker Specification (Rust)

## 1. Overview and Architectural Boundary

In accordance with `AGENTS.md` and `docs/3. ARCHITECTURE.md`:
- **Role**: Private delivery-dispatch optimization worker.
- **Strict Boundary**: Rust is strictly an offline computational worker. Rust **must not** own authentication, pricing, payment state, or arbitrary order transitions.
- **Go Authority**: Go owns public APIs, custom authentication, database schema, payment orchestration, and authoritative order state transitions. Go validates every proposal emitted by the dispatch worker before committing an assignment.
- **Current Status**: **DEFERRED**. Delivery simulation is currently handled natively in Go (`internal/delivery`), and `cargo` is not present in the local environment. When dispatch scaling and optimization demands warrant dedicated native execution, this specification governs the implementation.

---

## 2. Dispatch Proposal Protocol

### 2.1 Trigger / Event Ingestion
When an order reaches `status = "preparing"`, the Go backend emits a transactional outbox event:
- **EventType**: `dispatch.proposal_requested`
- **AggregateType**: `order`
- **Payload Schema**:

```json
{
  "orderId": "ord_12345678",
  "restaurant": {
    "id": "rest_01",
    "name": "Sultan's Dine",
    "latitude": 23.7937,
    "longitude": 90.4066,
    "zoneId": "zone_banani"
  },
  "destination": {
    "addressId": "addr_01",
    "latitude": 23.7925,
    "longitude": 90.4078
  },
  "readyAtEstimated": "2026-10-19T13:30:00Z",
  "candidateRiders": [
    {
      "riderId": "rdr_001",
      "latitude": 23.7940,
      "longitude": 90.4050,
      "activeOrdersCount": 0,
      "rating": 4.9
    },
    {
      "riderId": "rdr_002",
      "latitude": 23.8010,
      "longitude": 90.4120,
      "activeOrdersCount": 1,
      "rating": 4.7
    }
  ]
}
```

### 2.2 Optimization & Scoring Algorithm
The dispatch worker computes a weighted score for each candidate:
$$\text{Score} = w_1 \cdot \text{DistanceScore} + w_2 \cdot \text{WorkloadPenalty} + w_3 \cdot \text{RatingBonus}$$
- **Distance**: Haversine distance between rider's current coordinates and restaurant pickup.
- **Workload**: Penalizes riders currently carrying orders.
- **Availability Window**: Evaluates whether rider can reach restaurant before `readyAtEstimated`.

### 2.3 Proposal Output (Rust to Go)
The worker outputs an assignment proposal:
- **EventType**: `dispatch.proposal_emitted`
- **Payload Schema**:

```json
{
  "orderId": "ord_12345678",
  "algorithmVersion": "v1.0.0-greedy",
  "selectedRiderId": "rdr_001",
  "confidenceScore": 0.94,
  "estimatedPickupMinutes": 6,
  "estimatedDeliveryMinutes": 18,
  "rankedCandidates": [
    {
      "riderId": "rdr_001",
      "score": 0.94,
      "reasonCode": "CLOSEST_PROXIMITY_ZERO_LOAD"
    },
    {
      "riderId": "rdr_002",
      "score": 0.68,
      "reasonCode": "FURTHER_DISTANCE_ACTIVE_ORDER"
    }
  ],
  "evaluatedAt": "2026-10-19T13:10:00Z"
}
```

### 2.4 Authoritative Go Validation
Before committing:
1. Go locks the order record in PostgreSQL.
2. Go verifies the order is still in `preparing` or `accepted` status (not cancelled by customer).
3. Go verifies `selectedRiderId` is still online and has not been assigned another order in the intervening lease window.
4. Go commits the assignment transactionally, updates `order.rider_id = selectedRiderId`, and advances tracking.

---

## 3. Implementation Plan When Activated
When activated in `services/dispatch-rust/`:
1. `Cargo.toml` with dependencies: `serde`, `serde_json`, `tokio`, `sqlx` (read-only views), `geo`.
2. Clippy & formatting checks: `cargo fmt --check`, `cargo clippy --all-targets --all-features -- -D warnings`, `cargo test`.
3. Private DB worker lease model with optimistic concurrency control (`lease_expires_at`).
