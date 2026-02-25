package behavior_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"food_ordering_coordination_system/internal/domain"
	httpapi "food_ordering_coordination_system/internal/http"
	persistence "food_ordering_coordination_system/internal/persistance"
	"github.com/cucumber/godog"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type scenarioState struct {
	container         *mongodb.MongoDBContainer
	client            *mongo.Client
	db                *mongo.Database
	repo              *persistence.MongoRepository
	router            http.Handler
	token             string
	memberID          uuid.UUID
	response          *httptest.ResponseRecorder
	initialOrderCount int64
	initialEventCount int64
	invalidBody       map[string]any
}

func InitializeScenario(sc *godog.ScenarioContext) {
	state := &scenarioState{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, state.reset()
	})

	sc.After(func(_ context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		return context.Background(), state.close()
	})

	sc.Step(`^a member has ([0-9]+(?:\\.[0-9]+)?) credits$`, state.aMemberHasCredits)
	sc.Step(`^the member places an order totaling ([0-9]+(?:\\.[0-9]+)?)$`, state.theMemberPlacesAnOrderTotaling)
	sc.Step(`^the order is confirmed$`, state.theOrderIsConfirmed)
	sc.Step(`^remaining credits should be ([0-9]+(?:\\.[0-9]+)?)$`, state.remainingCreditsShouldBe)
	sc.Step(`^an order-created event exists$`, state.anOrderCreatedEventExists)
	sc.Step(`^the request is rejected with "([^"]*)"$`, state.theRequestIsRejectedWith)
	sc.Step(`^credits remain ([0-9]+(?:\\.[0-9]+)?)$`, state.creditsRemain)
	sc.Step(`^no order or event is created$`, state.noOrderOrEventIsCreated)
	sc.Step(`^a malformed order payload$`, state.aMalformedOrderPayload)
	sc.Step(`^the malformed order is submitted$`, state.theMalformedOrderIsSubmitted)
	sc.Step(`^the response status should be ([0-9]+)$`, state.theResponseStatusShouldBe)
	sc.Step(`^state is unchanged$`, state.stateIsUnchanged)
}

func (s *scenarioState) reset() error {
	ctx := context.Background()
	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		return err
	}
	s.container = container

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		return err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		return err
	}
	s.client = client

	dbName := "food_ordering_behavior_" + uuid.NewString()
	s.db = client.Database(dbName)
	s.repo = persistence.NewMongoRepository(s.db)
	if err := s.repo.EnsureSchema(ctx); err != nil {
		return err
	}

	service := domain.NewFoodOrderingService(s.repo, s.repo, s.repo)
	s.router = httpapi.NewFoodOrderingRouter(service, nil)

	s.memberID = uuid.Nil
	s.response = nil
	s.token = ""
	s.invalidBody = nil
	s.initialOrderCount = 0
	s.initialEventCount = 0
	return nil
}

func (s *scenarioState) close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if s.client != nil {
		_ = s.client.Disconnect(ctx)
	}
	if s.container != nil {
		return s.container.Terminate(ctx)
	}
	return nil
}

func (s *scenarioState) aMemberHasCredits(rawCredits string) error {
	credits, err := strconv.ParseFloat(rawCredits, 64)
	if err != nil {
		return err
	}

	s.memberID = uuid.New()
	s.token = tokenForSubject(s.memberID, httpapi.RoleMember)
	if err := s.repo.Set(s.memberID, credits); err != nil {
		return err
	}

	s.initialOrderCount, err = s.db.Collection("orders").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	s.initialEventCount, err = s.db.Collection("events").CountDocuments(context.Background(), bson.M{})
	return err
}

func (s *scenarioState) theMemberPlacesAnOrderTotaling(rawTotal string) error {
	total, err := strconv.ParseFloat(rawTotal, 64)
	if err != nil {
		return err
	}

	body := map[string]any{
		"member_id": s.memberID.String(),
		"items": []map[string]any{
			{"id": uuid.New().String(), "quantity": 1, "price": total},
		},
		"delivery_notes": "",
	}

	s.response = s.postOrder(body, s.token)
	return nil
}

func (s *scenarioState) theOrderIsConfirmed() error {
	if s.response == nil {
		return fmt.Errorf("request was not executed")
	}
	if s.response.Code != http.StatusCreated {
		return fmt.Errorf("expected status %d, got %d", http.StatusCreated, s.response.Code)
	}

	var response struct {
		Status string `json:"status"`
	}
	if err := decodeBody(s.response, &response); err != nil {
		return err
	}
	if response.Status != "CONFIRMED" {
		return fmt.Errorf("expected CONFIRMED status, got %s", response.Status)
	}
	return nil
}

func (s *scenarioState) remainingCreditsShouldBe(rawCredits string) error {
	expected, err := strconv.ParseFloat(rawCredits, 64)
	if err != nil {
		return err
	}

	if s.response == nil {
		return fmt.Errorf("request was not executed")
	}

	var response struct {
		RemainingCredits float64 `json:"remaining_credits"`
	}
	if err := decodeBody(s.response, &response); err != nil {
		return err
	}
	if response.RemainingCredits != expected {
		return fmt.Errorf("expected remaining credits %v, got %v", expected, response.RemainingCredits)
	}
	return nil
}

func (s *scenarioState) anOrderCreatedEventExists() error {
	count, err := s.db.Collection("events").CountDocuments(context.Background(), bson.M{"type": string(domain.FoodOrderCreatedEvt)})
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("expected 1 order-created event, got %d", count)
	}
	return nil
}

func (s *scenarioState) theRequestIsRejectedWith(code string) error {
	if s.response == nil {
		return fmt.Errorf("request was not executed")
	}
	if s.response.Code != http.StatusUnprocessableEntity {
		return fmt.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, s.response.Code)
	}

	var response struct {
		Code string `json:"code"`
	}
	if err := decodeBody(s.response, &response); err != nil {
		return err
	}
	if response.Code != code {
		return fmt.Errorf("expected error code %s, got %s", code, response.Code)
	}
	return nil
}

func (s *scenarioState) creditsRemain(rawCredits string) error {
	expected, err := strconv.ParseFloat(rawCredits, 64)
	if err != nil {
		return err
	}

	actual, ok, err := s.repo.Get(s.memberID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected credits account to exist")
	}
	if actual != expected {
		return fmt.Errorf("expected credits %v, got %v", expected, actual)
	}
	return nil
}

func (s *scenarioState) noOrderOrEventIsCreated() error {
	orderCount, err := s.db.Collection("orders").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	if orderCount != s.initialOrderCount {
		return fmt.Errorf("expected order count %d, got %d", s.initialOrderCount, orderCount)
	}

	eventCount, err := s.db.Collection("events").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	if eventCount != s.initialEventCount {
		return fmt.Errorf("expected event count %d, got %d", s.initialEventCount, eventCount)
	}
	return nil
}

func (s *scenarioState) aMalformedOrderPayload() error {
	adminSubject := uuid.New()
	s.token = tokenForSubject(adminSubject, httpapi.RoleHiveManager)

	var err error
	s.initialOrderCount, err = s.db.Collection("orders").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	s.initialEventCount, err = s.db.Collection("events").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	s.invalidBody = map[string]any{
		"member_id": "not-a-uuid",
		"items": []map[string]any{{
			"id":       uuid.New().String(),
			"quantity": 0,
			"price":    -2.0,
		}},
	}
	return nil
}

func (s *scenarioState) theMalformedOrderIsSubmitted() error {
	if s.invalidBody == nil {
		return fmt.Errorf("invalid body not configured")
	}
	s.response = s.postOrder(s.invalidBody, s.token)
	return nil
}

func (s *scenarioState) theResponseStatusShouldBe(rawStatus string) error {
	expected, err := strconv.Atoi(rawStatus)
	if err != nil {
		return err
	}
	if s.response == nil {
		return fmt.Errorf("request was not executed")
	}
	if s.response.Code != expected {
		return fmt.Errorf("expected status %d, got %d", expected, s.response.Code)
	}
	return nil
}

func (s *scenarioState) stateIsUnchanged() error {
	return s.noOrderOrEventIsCreated()
}

func (s *scenarioState) postOrder(body any, token string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	s.router.ServeHTTP(rec, req)
	return rec
}

func decodeBody(rec *httptest.ResponseRecorder, target any) error {
	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		return fmt.Errorf("decode body: %w", err)
	}
	return nil
}

func tokenForSubject(subject uuid.UUID, role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  subject.String(),
		"role": role,
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	})

	raw, _ := token.SignedString([]byte("test-signing-key"))
	return raw
}
