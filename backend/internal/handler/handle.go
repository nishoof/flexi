package handler

import (
	"net/http"

	"github.com/nishoof/flexi/backend/internal/apierr"
)

// apiHandler returns an *apierr.Error on failure, or nil on success.
// On success the handler writes the response itself.
type apiHandler func(http.ResponseWriter, *http.Request) *apierr.Error

// Handle turns an apiHandler into an http.HandlerFunc.
func Handle(h apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			err.WriteTo(w)
		}
	}
}
