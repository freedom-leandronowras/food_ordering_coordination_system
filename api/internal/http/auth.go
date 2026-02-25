package httpapi

import (
	"errors"
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
	Role    string
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Authenticator struct{}

type AuthenticatedHandler func(http.ResponseWriter, *http.Request, AuthClaims)

func NewAuthenticator() *Authenticator {
	return &Authenticator{}
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
	claims := &jwtClaims{}
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(rawToken, claims); err != nil {
		return AuthClaims{}, errors.New("invalid token")
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
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

	return AuthClaims{
		Subject: subjectID,
		Role:    role,
	}, nil
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
