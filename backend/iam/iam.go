package iam

import (
	"context"
	"strings"

	"flowdb/backend/policies"
	"flowdb/backend/store"
)

type Authorizer struct {
	store   *store.Store
	polices *policies.Store
}

type Decision struct {
	Allowed     bool
	Constraints policies.Constraints
}

func NewAuthorizer(st *store.Store, pol *policies.Store) *Authorizer {
	return &Authorizer{store: st, polices: pol}
}

func (a *Authorizer) Authorize(ctx context.Context, user store.User, action string, resource string, env string) (Decision, error) {
	if user.IsAdmin {
		return Decision{Allowed: true}, nil
	}
	roles, err := a.store.ListUserRoles(ctx, user.ID)
	if err != nil {
		return Decision{}, err
	}
	allowedByRole := false
	for _, role := range roles {
		if roleAllows(role.Permissions, action) {
			allowedByRole = true
			break
		}
	}
	if !allowedByRole {
		return Decision{Allowed: false}, nil
	}
	engine := a.polices.Engine()
	allowedByPolicy, constraints := engine.Evaluate(action, resource, env)
	if !engine.HasRules() {
		return Decision{Allowed: true}, nil
	}
	return Decision{Allowed: allowedByPolicy, Constraints: constraints}, nil
}

func roleAllows(perms []string, action string) bool {
	for _, p := range perms {
		if p == "*" || strings.EqualFold(p, action) {
			return true
		}
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(strings.ToLower(action), strings.ToLower(prefix)) {
				return true
			}
		}
	}
	return false
}
