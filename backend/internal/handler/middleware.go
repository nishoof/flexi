package handler

import (
	"context"
	"net/http"

	"github.com/nishoof/flexi/backend/internal/apierr"
	"github.com/nishoof/flexi/backend/internal/util"
)

type contextKey int

const userIDKey contextKey = iota

// withAuth checks auth, then calls next with the user id in context.
func withAuth(next apiHandler) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) *apierr.Error {
		userId, err := util.AuthenticateUser(r)
		if err != nil {
			return apierr.Unauthorized("Unauthorized")
		}
		ctx := context.WithValue(r.Context(), userIDKey, userId)
		return next(w, r.WithContext(ctx))
	}
}

func userID(r *http.Request) int64 {
	return r.Context().Value(userIDKey).(int64)
}
