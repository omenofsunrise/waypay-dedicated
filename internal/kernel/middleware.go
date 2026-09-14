package kernel

import (
	"context"
	"net/http"
	"waypay_dedicated/internal/rbac"
)

type ctxKey string

const userKey ctxKey = "user"

func (k *Kernel) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := "test_id"
		roles := []string{"admin"}
		perms := []string{"view.loans"}

		u := rbac.User{ID: userID, Roles: roles, Permissions: perms}
		ctx := context.WithValue(r.Context(), userKey, u)
		next(w, r.WithContext(ctx))
	}
}

func (k *Kernel) RequirePerm(perm string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := r.Context().Value(userKey).(rbac.User)
		if !k.rbac.Can(u, perm) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
