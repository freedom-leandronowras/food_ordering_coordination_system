package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	persistence "food_ordering_coordination_system/internal/persistance"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength = 8

var domainPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)+$`)

type AuthController struct {
	users                persistence.UserRepository
	credits              persistence.CreditRepository
	authenticator        *Authenticator
	tokenTTL             time.Duration
	allowSelfAssignRoles bool
}

func NewAuthController(
	users persistence.UserRepository,
	credits persistence.CreditRepository,
	authenticator *Authenticator,
	tokenTTL time.Duration,
	allowSelfAssignRoles bool,
) *AuthController {
	if tokenTTL <= 0 {
		tokenTTL = time.Hour
	}
	return &AuthController{
		users:                users,
		credits:              credits,
		authenticator:        authenticator,
		tokenTTL:             tokenTTL,
		allowSelfAssignRoles: allowSelfAssignRoles,
	}
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	if c == nil || c.users == nil || c.authenticator == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "authentication is not configured")
		return
	}

	var req registerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "request body is invalid")
		return
	}

	email := normalizeEmailAddress(req.Email)
	if !isValidEmail(email) {
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", "email is invalid")
		return
	}

	if len(req.Password) < minPasswordLength {
		writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", "password must be at least 8 characters")
		return
	}

	role := normalizeRoleValue(req.Role)
	if role == "" {
		role = RoleMember
	}
	if !isSupportedRole(role) {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE", "role is invalid")
		return
	}
	if role != RoleMember && !c.allowSelfAssignRoles {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "role assignment is restricted")
		return
	}

	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		fullName = "Member"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create account")
		return
	}

	user := persistence.User{
		UserID:       uuid.New(),
		MemberID:     uuid.New(),
		Email:        email,
		FullName:     fullName,
		Role:         role,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := c.users.CreateUser(user); err != nil {
		if isDuplicateKeyError(err) {
			writeError(w, http.StatusConflict, "ACCOUNT_EXISTS", "an account already exists for this email")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create account")
		return
	}

	token, err := c.authenticator.IssueToken(user.MemberID, user.UserID, user.Email, user.Role, c.tokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create session")
		return
	}

	credits := c.loadCredits(user.MemberID)
	writeJSON(w, http.StatusCreated, authSessionResponse{
		Token: token,
		User:  toAuthUserResponse(user, credits),
	})
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if c == nil || c.users == nil || c.authenticator == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "authentication is not configured")
		return
	}

	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "request body is invalid")
		return
	}

	email := normalizeEmailAddress(req.Email)
	if !isValidEmail(email) {
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", "email is invalid")
		return
	}

	user, found, err := c.users.FindUserByEmail(email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not authenticate user")
		return
	}
	if !found {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
		return
	}

	token, err := c.authenticator.IssueToken(user.MemberID, user.UserID, user.Email, user.Role, c.tokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create session")
		return
	}

	credits := c.loadCredits(user.MemberID)
	writeJSON(w, http.StatusOK, authSessionResponse{
		Token: token,
		User:  toAuthUserResponse(user, credits),
	})
}

func (c *AuthController) Me(w http.ResponseWriter, _ *http.Request, auth AuthClaims) {
	if c == nil || c.users == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "authentication is not configured")
		return
	}

	user, found, err := c.users.FindUserByMemberID(auth.Subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch member profile")
		return
	}
	if !found {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "member profile not found")
		return
	}

	credits := c.loadCredits(user.MemberID)
	writeJSON(w, http.StatusOK, toAuthUserResponse(user, credits))
}

func (c *AuthController) ListMembersByDomain(w http.ResponseWriter, r *http.Request, _ AuthClaims) {
	if c == nil || c.users == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "authentication is not configured")
		return
	}

	domain := normalizeDomainValue(r.URL.Query().Get("domain"))
	if domain != "" && !domainPattern.MatchString(domain) {
		writeError(w, http.StatusBadRequest, "INVALID_DOMAIN", "domain query is invalid")
		return
	}

	var (
		users []persistence.User
		err   error
	)
	if domain == "" {
		users, err = c.users.ListUsers()
	} else {
		users, err = c.users.ListUsersByEmailDomain(domain)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not query members")
		return
	}

	members := make([]authUserResponse, 0, len(users))
	for _, user := range users {
		members = append(members, toAuthUserResponse(user, c.loadCredits(user.MemberID)))
	}

	writeJSON(w, http.StatusOK, membersByDomainResponse{
		Domain:  domain,
		Members: members,
	})
}

func normalizeEmailAddress(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeDomainValue(domain string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(domain, "@")))
}

func normalizeRoleValue(role string) string {
	return strings.ToUpper(strings.TrimSpace(role))
}

func isValidEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return normalizeEmailAddress(parsed.Address) == email
}

func isSupportedRole(role string) bool {
	switch role {
	case RoleMember, RoleHiveManager, RoleInnovationLead:
		return true
	default:
		return false
	}
}

func toAuthUserResponse(user persistence.User, credits float64) authUserResponse {
	return authUserResponse{
		UserID:   user.UserID.String(),
		MemberID: user.MemberID.String(),
		Email:    normalizeEmailAddress(user.Email),
		FullName: strings.TrimSpace(user.FullName),
		Role:     normalizeRoleValue(user.Role),
		Credits:  credits,
	}
}

func (c *AuthController) loadCredits(memberID uuid.UUID) float64 {
	if c == nil || c.credits == nil {
		return 0
	}

	credits, found, err := c.credits.Get(memberID)
	if err != nil || !found {
		return 0
	}
	return credits
}

func isDuplicateKeyError(err error) bool {
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, e := range writeErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}

	var bulkErr mongo.BulkWriteException
	if errors.As(err, &bulkErr) {
		for _, e := range bulkErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}

	return false
}
