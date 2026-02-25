package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"handler/internal/domain"
	httpapi "handler/internal/http"
	persistence "handler/internal/persistance"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestPostOrderSuccessAndCreditsDeduction(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	if err := env.repo.Set(memberID, 100); err != nil {
		t.Fatalf("seed credits: %v", err)
	}

	body := map[string]any{
		"member_id": memberID.String(),
		"items": []map[string]any{
			{"id": uuid.New().String(), "quantity": 2, "price": 15.0},
		},
		"delivery_notes": "no onions",
	}

	rec := executeJSONRequest(t, env.router, http.MethodPost, "/api/orders", body, tokenFor(t, memberID, httpapi.RoleMember))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response struct {
		OrderID          string  `json:"order_id"`
		Status           string  `json:"status"`
		TotalPrice       float64 `json:"total_price"`
		RemainingCredits float64 `json:"remaining_credits"`
	}
	decodeResponse(t, rec, &response)

	if response.OrderID == "" {
		t.Fatal("expected non-empty order_id")
	}
	if response.Status != "CONFIRMED" {
		t.Fatalf("expected status CONFIRMED, got %s", response.Status)
	}
	if response.TotalPrice != 30 {
		t.Fatalf("expected total_price 30, got %v", response.TotalPrice)
	}
	if response.RemainingCredits != 70 {
		t.Fatalf("expected remaining_credits 70, got %v", response.RemainingCredits)
	}

	creditsRec := httptest.NewRecorder()
	creditsReq := httptest.NewRequest(http.MethodGet, "/api/members/"+memberID.String()+"/credits", nil)
	creditsReq.Header.Set("Authorization", "Bearer "+tokenFor(t, memberID, httpapi.RoleMember))
	env.router.ServeHTTP(creditsRec, creditsReq)

	if creditsRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, creditsRec.Code, creditsRec.Body.String())
	}
	var creditsResponse struct {
		Credits float64 `json:"credits"`
	}
	decodeResponse(t, creditsRec, &creditsResponse)
	if creditsResponse.Credits != 70 {
		t.Fatalf("expected credits 70, got %v", creditsResponse.Credits)
	}

	assertCollectionCount(t, env.db, "orders", 1)
	assertCollectionCount(t, env.db, "events", 1)

	var eventDoc struct {
		Type string `bson:"type"`
	}
	if err := env.db.Collection("events").FindOne(context.Background(), bson.M{}).Decode(&eventDoc); err != nil {
		t.Fatalf("find event: %v", err)
	}
	if eventDoc.Type != string(domain.FoodOrderCreatedEvt) {
		t.Fatalf("expected event type %s, got %s", domain.FoodOrderCreatedEvt, eventDoc.Type)
	}
}

func TestPostOrderInsufficientCredits(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	if err := env.repo.Set(memberID, 10); err != nil {
		t.Fatalf("seed credits: %v", err)
	}

	body := map[string]any{
		"member_id": memberID.String(),
		"items": []map[string]any{
			{"id": uuid.New().String(), "quantity": 1, "price": 25.0},
		},
		"delivery_notes": "",
	}

	rec := executeJSONRequest(t, env.router, http.MethodPost, "/api/orders", body, tokenFor(t, memberID, httpapi.RoleMember))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	}

	var response struct {
		Code string `json:"code"`
	}
	decodeResponse(t, rec, &response)
	if response.Code != "INSUFFICIENT_CREDITS" {
		t.Fatalf("expected code INSUFFICIENT_CREDITS, got %s", response.Code)
	}

	credits, ok, err := env.repo.Get(memberID)
	if err != nil {
		t.Fatalf("read credits: %v", err)
	}
	if !ok {
		t.Fatal("expected credits account to exist")
	}
	if credits != 10 {
		t.Fatalf("expected credits to remain 10, got %v", credits)
	}

	assertCollectionCount(t, env.db, "orders", 0)
	assertCollectionCount(t, env.db, "events", 0)
}

func TestPostOrderInvalidPayload(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	token := tokenFor(t, memberID, httpapi.RoleMember)
	testCases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "invalid member id",
			body: map[string]any{
				"member_id": "not-a-uuid",
				"items": []map[string]any{
					{"id": uuid.New().String(), "quantity": 1, "price": 5.0},
				},
			},
		},
		{
			name: "invalid quantity",
			body: map[string]any{
				"member_id": memberID.String(),
				"items": []map[string]any{
					{"id": uuid.New().String(), "quantity": 0, "price": 5.0},
				},
			},
		},
		{
			name: "negative price",
			body: map[string]any{
				"member_id": memberID.String(),
				"items": []map[string]any{
					{"id": uuid.New().String(), "quantity": 1, "price": -1.0},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := executeJSONRequest(t, env.router, http.MethodPost, "/api/orders", tc.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})
	}

	assertCollectionCount(t, env.db, "orders", 0)
	assertCollectionCount(t, env.db, "events", 0)
}

func TestRbacMemberCannotReadOtherMemberCredits(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	requesterID := uuid.New()
	targetID := uuid.New()
	if err := env.repo.Set(targetID, 42); err != nil {
		t.Fatalf("seed target credits: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/members/"+targetID.String()+"/credits", nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, requesterID, httpapi.RoleMember))
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestRbacManagerCanReadAnyMemberCredits(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	targetID := uuid.New()
	if err := env.repo.Set(targetID, 42); err != nil {
		t.Fatalf("seed target credits: %v", err)
	}

	managerRoles := []string{
		httpapi.RoleHiveManager,
		httpapi.RoleInnovationLead,
	}

	for _, role := range managerRoles {
		t.Run(role, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/members/"+targetID.String()+"/credits", nil)
			req.Header.Set("Authorization", "Bearer "+tokenFor(t, uuid.New(), role))
			env.router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreditsCapRejectsAmountsAbove1000(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	err := env.repo.Set(memberID, 1001)
	if err == nil {
		t.Fatal("expected error when setting credits above cap")
	}

	assertCollectionCount(t, env.db, "credits", 0)
}

type testEnv struct {
	container *mongodb.MongoDBContainer
	client    *mongo.Client
	db        *mongo.Database
	repo      *persistence.MongoRepository
	router    http.Handler
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	ctx := context.Background()
	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongodb container: %v", err)
	}

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get mongodb connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		t.Fatalf("connect mongodb: %v", err)
	}

	dbName := "food_ordering_test_" + uuid.NewString()
	db := client.Database(dbName)

	repo := persistence.NewMongoRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	service := domain.NewFoodOrderingService(repo, repo, repo)
	router := httpapi.NewFoodOrderingRouter(service, nil)

	return &testEnv{
		container: container,
		client:    client,
		db:        db,
		repo:      repo,
		router:    router,
	}
}

func (e *testEnv) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.client.Disconnect(ctx)
	_ = e.container.Terminate(ctx)
}

func tokenFor(t *testing.T, subject uuid.UUID, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  subject.String(),
		"role": role,
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	})

	raw, err := token.SignedString([]byte("test-signing-key"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return raw
}

func executeJSONRequest(t *testing.T, router http.Handler, method string, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response: %v body=%s", err, rec.Body.String())
	}
}

func assertCollectionCount(t *testing.T, db *mongo.Database, name string, expected int64) {
	t.Helper()
	count, err := db.Collection(name).CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("count %s: %v", name, err)
	}
	if count != expected {
		t.Fatalf("expected %d documents in %s, got %d", expected, name, count)
	}
}
