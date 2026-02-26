package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"food_ordering_coordination_system/internal/domain"
	httpapi "food_ordering_coordination_system/internal/http"
)

const testJWTSigningKey = "test-signing-key"

func TestAuthRegisterLoginAndMe(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	router := newAuthRouterForEnv(env, false)

	registerBody := map[string]any{
		"email":     "member@example.com",
		"password":  "password123",
		"full_name": "Example Member",
	}
	registerRec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/register", registerBody)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d body=%s", http.StatusCreated, registerRec.Code, registerRec.Body.String())
	}

	var registerResp struct {
		Token string `json:"token"`
		User  struct {
			MemberID string `json:"member_id"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		} `json:"user"`
	}
	decodeResponse(t, registerRec, &registerResp)
	if registerResp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if registerResp.User.Role != httpapi.RoleMember {
		t.Fatalf("expected role %s, got %s", httpapi.RoleMember, registerResp.User.Role)
	}
	if registerResp.User.Email != "member@example.com" {
		t.Fatalf("expected normalized email, got %s", registerResp.User.Email)
	}

	meRec := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+registerResp.Token)
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, meRec.Code, meRec.Body.String())
	}

	var meResp struct {
		MemberID string `json:"member_id"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}
	decodeResponse(t, meRec, &meResp)
	if meResp.MemberID != registerResp.User.MemberID {
		t.Fatalf("expected member id %s, got %s", registerResp.User.MemberID, meResp.MemberID)
	}
	if meResp.Email != "member@example.com" {
		t.Fatalf("expected member@example.com, got %s", meResp.Email)
	}

	loginRec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    "member@example.com",
		"password": "password123",
	})
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, loginRec.Code, loginRec.Body.String())
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	decodeResponse(t, loginRec, &loginResp)
	if loginResp.Token == "" {
		t.Fatal("expected non-empty token from login")
	}
}

func TestAuthRegisterManagerRoleBlockedWhenSelfAssignDisabled(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	router := newAuthRouterForEnv(env, false)

	rec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/register", map[string]any{
		"email":     "manager@example.com",
		"password":  "password123",
		"full_name": "Example Manager",
		"role":      httpapi.RoleHiveManager,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestAuthManagerCanLookupMembersByDomain(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	router := newAuthRouterForEnv(env, true)

	managerRec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/register", map[string]any{
		"email":     "manager@acme.io",
		"password":  "password123",
		"full_name": "Manager",
		"role":      httpapi.RoleHiveManager,
	})
	if managerRec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d body=%s", http.StatusCreated, managerRec.Code, managerRec.Body.String())
	}

	var managerResp struct {
		Token string `json:"token"`
	}
	decodeResponse(t, managerRec, &managerResp)

	for _, email := range []string{"alice@acme.io", "bob@acme.io", "user@other.io"} {
		rec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/register", map[string]any{
			"email":     email,
			"password":  "password123",
			"full_name": "Member",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("register %s failed: status=%d body=%s", email, rec.Code, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/members?domain=acme.io", nil)
	req.Header.Set("Authorization", "Bearer "+managerResp.Token)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Domain  string `json:"domain"`
		Members []struct {
			Email string `json:"email"`
		} `json:"members"`
	}
	decodeResponse(t, rec, &response)
	if response.Domain != "acme.io" {
		t.Fatalf("expected domain acme.io, got %s", response.Domain)
	}
	if len(response.Members) != 3 {
		t.Fatalf("expected 3 members in acme.io domain, got %d", len(response.Members))
	}
}

func TestAuthManagerCanListAllMembersWithoutDomainFilter(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	router := newAuthRouterForEnv(env, true)

	managerRec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/register", map[string]any{
		"email":     "lead@acme.io",
		"password":  "password123",
		"full_name": "Lead",
		"role":      httpapi.RoleInnovationLead,
	})
	if managerRec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d body=%s", http.StatusCreated, managerRec.Code, managerRec.Body.String())
	}

	var managerResp struct {
		Token string `json:"token"`
	}
	decodeResponse(t, managerRec, &managerResp)

	for _, email := range []string{"alice@acme.io", "bob@acme.io", "user@other.io"} {
		rec := executeJSONRequestNoAuth(t, router, http.MethodPost, "/api/auth/register", map[string]any{
			"email":     email,
			"password":  "password123",
			"full_name": "Member",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("register %s failed: status=%d body=%s", email, rec.Code, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/members", nil)
	req.Header.Set("Authorization", "Bearer "+managerResp.Token)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Domain  string `json:"domain"`
		Members []struct {
			Email string `json:"email"`
		} `json:"members"`
	}
	decodeResponse(t, rec, &response)
	if response.Domain != "" {
		t.Fatalf("expected empty domain for unfiltered list, got %q", response.Domain)
	}
	if len(response.Members) != 4 {
		t.Fatalf("expected 4 members total, got %d", len(response.Members))
	}
}

func newAuthRouterForEnv(env *testEnv, allowSelfAssignRoles bool) http.Handler {
	service := domain.NewFoodOrderingService(env.repo, env.repo, env.repo)
	authenticator := httpapi.NewAuthenticator(testJWTSigningKey)
	authController := httpapi.NewAuthController(env.repo, env.repo, authenticator, time.Hour, allowSelfAssignRoles)
	return httpapi.NewFoodOrderingRouterWithAuth(service, nil, authController, testJWTSigningKey)
}

func executeJSONRequestNoAuth(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body any,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	return rec
}
