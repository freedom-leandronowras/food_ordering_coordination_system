package persistence_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"food_ordering_coordination_system/internal/domain"
	persistence "food_ordering_coordination_system/internal/persistance"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// 1. Full-field roundtrip — every field that goes in must come back out.
//    Without this the event log is lossy and replay produces wrong state.
// ---------------------------------------------------------------------------

func TestAppendAndRetrieve_AllFieldsSurviveRoundtrip(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	aggregateID := uuid.New()
	correlationID := uuid.New()
	causationID := uuid.New()
	eventID := uuid.New()
	occurredAt := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	original := domain.Event{
		ID:            eventID,
		Type:          domain.FoodOrderCreatedEvt,
		AggregateID:   aggregateID,
		CorrelationID: correlationID,
		CausationID:   causationID,
		OccurredAt:    occurredAt,
		Payload: map[string]any{
			"order_id":    uuid.New().String(),
			"member_id":   uuid.New().String(),
			"total_price": 42.5,
			"status":      "CONFIRMED",
		},
	}

	if err := env.repo.Append(original); err != nil {
		t.Fatalf("append: %v", err)
	}

	events, err := env.repo.EventsByAggregate(context.Background(), aggregateID)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	got := events[0]

	if got.ID != eventID {
		t.Fatalf("ID: want %s, got %s", eventID, got.ID)
	}
	if got.Type != domain.FoodOrderCreatedEvt {
		t.Fatalf("Type: want %s, got %s", domain.FoodOrderCreatedEvt, got.Type)
	}
	if got.AggregateID != aggregateID {
		t.Fatalf("AggregateID: want %s, got %s", aggregateID, got.AggregateID)
	}
	if got.CorrelationID != correlationID {
		t.Fatalf("CorrelationID: want %s, got %s", correlationID, got.CorrelationID)
	}
	if got.CausationID != causationID {
		t.Fatalf("CausationID: want %s, got %s", causationID, got.CausationID)
	}
	if !got.OccurredAt.Equal(occurredAt) {
		t.Fatalf("OccurredAt: want %v, got %v", occurredAt, got.OccurredAt)
	}

	payload, ok := got.Payload.(map[string]any)
	if !ok {
		t.Fatalf("Payload type: want map[string]any, got %T", got.Payload)
	}
	if payload["total_price"] != 42.5 {
		t.Fatalf("Payload total_price: want 42.5, got %v", payload["total_price"])
	}
	if payload["status"] != "CONFIRMED" {
		t.Fatalf("Payload status: want CONFIRMED, got %v", payload["status"])
	}
}

// ---------------------------------------------------------------------------
// 2. Chronological ordering — replay MUST process events in occurred_at
//    order.  Insert out of order and verify they come back sorted.
// ---------------------------------------------------------------------------

func TestEventsByAggregate_ReturnsChronologicalOrder(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	aggregateID := uuid.New()
	base := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	// Append deliberately out of order: third, first, second.
	timestamps := []time.Time{
		base.Add(2 * time.Hour),
		base,
		base.Add(1 * time.Hour),
	}
	for i, ts := range timestamps {
		evt := domain.Event{
			ID:          uuid.New(),
			Type:        domain.EventType("test.ordering.v1"),
			AggregateID: aggregateID,
			OccurredAt:  ts,
			Payload:     map[string]any{"seq": int32(i)},
		}
		if err := env.repo.Append(evt); err != nil {
			t.Fatalf("append event %d: %v", i, err)
		}
	}

	events, err := env.repo.EventsByAggregate(context.Background(), aggregateID)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	for i := 1; i < len(events); i++ {
		if events[i].OccurredAt.Before(events[i-1].OccurredAt) {
			t.Fatalf("event %d (%v) is before event %d (%v) — not chronological",
				i, events[i].OccurredAt, i-1, events[i-1].OccurredAt)
		}
	}

	// The middle insert (base) must be first after sort.
	if !events[0].OccurredAt.Equal(base) {
		t.Fatalf("first event should be at %v, got %v", base, events[0].OccurredAt)
	}
}

// ---------------------------------------------------------------------------
// 3. Idempotency — the unique index on event ID must reject duplicates.
//    This is the safety net for at-least-once producers.
// ---------------------------------------------------------------------------

func TestAppend_RejectsDuplicateEventID(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	eventID := uuid.New()
	evt := domain.Event{
		ID:          eventID,
		Type:        domain.FoodOrderCreatedEvt,
		AggregateID: uuid.New(),
		OccurredAt:  time.Now().UTC(),
		Payload:     map[string]any{"attempt": "first"},
	}

	if err := env.repo.Append(evt); err != nil {
		t.Fatalf("first append: %v", err)
	}

	evt.Payload = map[string]any{"attempt": "second"}
	err := env.repo.Append(evt)
	if err == nil {
		t.Fatal("expected duplicate to be rejected, but append succeeded")
	}

	// Verify only the first event was stored.
	events, readErr := env.repo.EventsByAggregate(context.Background(), evt.AggregateID)
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}
	if len(events) != 1 {
		t.Fatalf("expected exactly 1 event after duplicate rejection, got %d", len(events))
	}
}

// ---------------------------------------------------------------------------
// 4. State derivation from replay — the defining guarantee of event
//    sourcing.  Persist a series of credit events and prove we can
//    reconstruct the correct balance by folding over the stream.
// ---------------------------------------------------------------------------

func TestDeriveStateByReplayingEvents(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	memberAggregateID := uuid.New()
	base := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)

	stream := []domain.Event{
		{
			ID: uuid.New(), Type: "credits.granted.v1",
			AggregateID: memberAggregateID, OccurredAt: base,
			Payload: map[string]any{"amount": 100.0, "reason": "monthly top-up"},
		},
		{
			ID: uuid.New(), Type: "credits.deducted.v1",
			AggregateID: memberAggregateID, OccurredAt: base.Add(1 * time.Hour),
			Payload: map[string]any{"amount": 30.0, "order_id": uuid.New().String()},
		},
		{
			ID: uuid.New(), Type: "credits.deducted.v1",
			AggregateID: memberAggregateID, OccurredAt: base.Add(2 * time.Hour),
			Payload: map[string]any{"amount": 15.0, "order_id": uuid.New().String()},
		},
		{
			ID: uuid.New(), Type: "credits.granted.v1",
			AggregateID: memberAggregateID, OccurredAt: base.Add(3 * time.Hour),
			Payload: map[string]any{"amount": 50.0, "reason": "manager top-up"},
		},
	}

	for i, evt := range stream {
		if err := env.repo.Append(evt); err != nil {
			t.Fatalf("append event %d: %v", i, err)
		}
	}

	events, err := env.repo.EventsByAggregate(context.Background(), memberAggregateID)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	// Fold: replay events to derive current balance.
	balance := 0.0
	for _, evt := range events {
		payload := evt.Payload.(map[string]any)
		amount := payload["amount"].(float64)
		switch evt.Type {
		case "credits.granted.v1":
			balance += amount
		case "credits.deducted.v1":
			balance -= amount
		}
	}

	// 100 − 30 − 15 + 50 = 105
	expected := 105.0
	if balance != expected {
		t.Fatalf("replayed balance: want %.2f, got %.2f", expected, balance)
	}
}

// ---------------------------------------------------------------------------
// 5. Aggregate stream isolation — events from one aggregate must never
//    leak into another's stream.  A broken query here silently corrupts
//    every replay in the system.
// ---------------------------------------------------------------------------

func TestEventsByAggregate_StreamsAreIsolated(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	orderA := uuid.New()
	orderB := uuid.New()

	for i := 0; i < 3; i++ {
		if err := env.repo.Append(domain.Event{
			ID: uuid.New(), Type: domain.FoodOrderCreatedEvt,
			AggregateID: orderA, OccurredAt: time.Now().UTC(),
			Payload: map[string]any{"aggregate": "A", "seq": int32(i)},
		}); err != nil {
			t.Fatalf("append A-%d: %v", i, err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := env.repo.Append(domain.Event{
			ID: uuid.New(), Type: domain.FoodOrderSubmittedEvt,
			AggregateID: orderB, OccurredAt: time.Now().UTC(),
			Payload: map[string]any{"aggregate": "B", "seq": int32(i)},
		}); err != nil {
			t.Fatalf("append B-%d: %v", i, err)
		}
	}

	eventsA, err := env.repo.EventsByAggregate(context.Background(), orderA)
	if err != nil {
		t.Fatalf("read A: %v", err)
	}
	eventsB, err := env.repo.EventsByAggregate(context.Background(), orderB)
	if err != nil {
		t.Fatalf("read B: %v", err)
	}

	if len(eventsA) != 3 {
		t.Fatalf("aggregate A: want 3 events, got %d", len(eventsA))
	}
	if len(eventsB) != 2 {
		t.Fatalf("aggregate B: want 2 events, got %d", len(eventsB))
	}

	for _, e := range eventsA {
		if e.AggregateID != orderA {
			t.Fatalf("aggregate A stream contains event from %s", e.AggregateID)
		}
	}
	for _, e := range eventsB {
		if e.AggregateID != orderB {
			t.Fatalf("aggregate B stream contains event from %s", e.AggregateID)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. Correlation / causation chain — when an order triggers a credit
//    deduction and an event, the chain must survive the roundtrip so
//    you can trace "why did this happen?".
// ---------------------------------------------------------------------------

func TestCorrelationAndCausationChain_SurvivesRoundtrip(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	aggregateID := uuid.New()
	correlationID := uuid.New()
	base := time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)

	// Event 1: order placed (root cause, no causation).
	orderPlaced := domain.Event{
		ID: uuid.New(), Type: domain.FoodOrderCreatedEvt,
		AggregateID: aggregateID, CorrelationID: correlationID,
		OccurredAt: base,
		Payload:    map[string]any{"step": "order-placed"},
	}

	// Event 2: credits deducted (caused by event 1).
	creditsDeducted := domain.Event{
		ID: uuid.New(), Type: "credits.deducted.v1",
		AggregateID: aggregateID, CorrelationID: correlationID,
		CausationID: orderPlaced.ID,
		OccurredAt:  base.Add(1 * time.Millisecond),
		Payload:     map[string]any{"step": "credits-deducted"},
	}

	// Event 3: order submitted to vendor (caused by event 2).
	orderSubmitted := domain.Event{
		ID: uuid.New(), Type: domain.FoodOrderSubmittedEvt,
		AggregateID: aggregateID, CorrelationID: correlationID,
		CausationID: creditsDeducted.ID,
		OccurredAt:  base.Add(2 * time.Millisecond),
		Payload:     map[string]any{"step": "order-submitted"},
	}

	for _, evt := range []domain.Event{orderPlaced, creditsDeducted, orderSubmitted} {
		if err := env.repo.Append(evt); err != nil {
			t.Fatalf("append: %v", err)
		}
	}

	events, err := env.repo.EventsByAggregate(context.Background(), aggregateID)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// All three share the same correlation ID.
	for _, e := range events {
		if e.CorrelationID != correlationID {
			t.Fatalf("event %s has correlation %s, want %s", e.ID, e.CorrelationID, correlationID)
		}
	}

	// Event 1 (root): no causation.
	if events[0].CausationID != uuid.Nil {
		t.Fatalf("root event should have nil causation, got %s", events[0].CausationID)
	}
	// Event 2: caused by event 1.
	if events[1].CausationID != orderPlaced.ID {
		t.Fatalf("event 2 causation: want %s, got %s", orderPlaced.ID, events[1].CausationID)
	}
	// Event 3: caused by event 2.
	if events[2].CausationID != creditsDeducted.ID {
		t.Fatalf("event 3 causation: want %s, got %s", creditsDeducted.ID, events[2].CausationID)
	}
}

// ---------------------------------------------------------------------------
// 7. Concurrent appends — multiple goroutines writing to the same
//    aggregate must not lose events or corrupt the stream.
// ---------------------------------------------------------------------------

func TestAppend_ConcurrentWritesToSameAggregate(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	aggregateID := uuid.New()
	const writers = 20

	var wg sync.WaitGroup
	errs := make(chan error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			evt := domain.Event{
				ID:          uuid.New(),
				Type:        domain.EventType("concurrent.write.v1"),
				AggregateID: aggregateID,
				OccurredAt:  time.Now().UTC(),
				Payload:     map[string]any{"writer": int32(seq)},
			}
			if err := env.repo.Append(evt); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent append failed: %v", err)
	}

	events, err := env.repo.EventsByAggregate(context.Background(), aggregateID)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(events) != writers {
		t.Fatalf("expected %d events, got %d", writers, len(events))
	}
}

// ---------------------------------------------------------------------------
// 8. Empty stream — a non-existent aggregate returns an empty slice,
//    not an error.  Callers must be able to distinguish "never existed"
//    from "something went wrong".
// ---------------------------------------------------------------------------

func TestEventsByAggregate_EmptyStreamReturnsEmptySlice(t *testing.T) {
	env := newEventStoreEnv(t)
	defer env.close()

	events, err := env.repo.EventsByAggregate(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events != nil {
		t.Fatalf("expected nil slice for unknown aggregate, got %d events", len(events))
	}
}

// ===========================================================================
// Test environment — one MongoDB container per test, same pattern as
// the existing router_test.go.
// ===========================================================================

type eventStoreEnv struct {
	container *mongodb.MongoDBContainer
	client    *mongo.Client
	repo      *persistence.MongoRepository
}

func newEventStoreEnv(t *testing.T) *eventStoreEnv {
	t.Helper()

	ctx := context.Background()
	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongodb container: %v", err)
	}

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		t.Fatalf("connect mongodb: %v", err)
	}

	db := client.Database("event_store_test_" + uuid.NewString())
	repo := persistence.NewMongoRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	return &eventStoreEnv{container: container, client: client, repo: repo}
}

func (e *eventStoreEnv) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.client.Disconnect(ctx)
	_ = e.container.Terminate(ctx)
}
