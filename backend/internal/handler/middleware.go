package handler

import (
	"context"
	"net/http"

	"github.com/nishoof/flexi/backend/internal/util"
)

type contextKey int

const userIDKey contextKey = iota

// withAuth returns a handler function that handles auth then calls the given
// handler function.
//
// If the request is not from an authenticated user, the handler writes a 401
// and returns.
//
// If the request is from an authenticated user, the handler creates a context
// with the userId and calls the given handler function with that context.
func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := util.AuthenticateUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userId)
		next(w, r.WithContext(ctx))
	}
}

func userID(r *http.Request) int64 {
	return r.Context().Value(userIDKey).(int64)
}
