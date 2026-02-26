package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	RoleMember         = "MEMBER"
	RoleHiveManager    = "HIVE_MANAGER"
	RoleInnovationLead = "INNOVATION_LEAD"
)

type AuthClaims struct {
	Subject uuid.UUID
	UserID  uuid.UUID
	Email   string
	Role    string
}

type jwtClaims struct {
	Role   string `json:"role"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

const DefaultJWTSigningKey = "test-signing-key"

type Authenticator struct {
	signingKey []byte
}

type AuthenticatedHandler func(http.ResponseWriter, *http.Request, AuthClaims)

func NewAuthenticator(signingKey string) *Authenticator {
	return &Authenticator{signingKey: []byte(strings.TrimSpace(signingKey))}
}

func (a *Authenticator) RequireRoles(roles ...string) func(AuthenticatedHandler) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next AuthenticatedHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			info, err := a.parseToken(r)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "token is missing or invalid")
				return
			}

			if _, ok := allowed[info.Role]; !ok {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "role does not have access to this endpoint")
				return
			}

			next(w, r, info)
		}
	}
}

func (a *Authenticator) parseToken(r *http.Request) (AuthClaims, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return AuthClaims{}, errors.New("missing bearer token")
	}

	rawToken := strings.TrimSpace(header[len("Bearer "):])
	if len(a.signingKey) == 0 {
		return AuthClaims{}, errors.New("missing signing key")
	}
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return a.signingKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return AuthClaims{}, errors.New("invalid token")
	}

	if claims.ExpiresAt == nil {
		return AuthClaims{}, errors.New("missing exp")
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return AuthClaims{}, errors.New("token expired")
	}

	subjectID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return AuthClaims{}, errors.New("invalid subject")
	}

	role := strings.ToUpper(strings.TrimSpace(claims.Role))
	if role == "" {
		return AuthClaims{}, errors.New("missing role")
	}

	var userID uuid.UUID
	if strings.TrimSpace(claims.UserID) != "" {
		parsedUserID, err := uuid.Parse(strings.TrimSpace(claims.UserID))
		if err != nil {
			return AuthClaims{}, errors.New("invalid user id")
		}
		userID = parsedUserID
	}

	return AuthClaims{
		Subject: subjectID,
		UserID:  userID,
		Email:   strings.ToLower(strings.TrimSpace(claims.Email)),
		Role:    role,
	}, nil
}

func (a *Authenticator) IssueToken(subject uuid.UUID, userID uuid.UUID, email string, role string, ttl time.Duration) (string, error) {
	if len(a.signingKey) == 0 {
		return "", errors.New("missing signing key")
	}
	if subject == uuid.Nil {
		return "", errors.New("subject is required")
	}
	normalizedRole := strings.ToUpper(strings.TrimSpace(role))
	if normalizedRole == "" {
		return "", errors.New("role is required")
	}
	if ttl <= 0 {
		ttl = time.Hour
	}

	now := time.Now().UTC()
	claims := jwtClaims{
		Role:   normalizedRole,
		UserID: userID.String(),
		Email:  strings.ToLower(strings.TrimSpace(email)),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.signingKey)
}

func canAccessMemberData(info AuthClaims, memberID uuid.UUID) bool {
	if isManagerRole(info.Role) {
		return true
	}
	return info.Role == RoleMember && info.Subject == memberID
}

func isManagerRole(role string) bool {
	return role == RoleHiveManager || role == RoleInnovationLead
}
