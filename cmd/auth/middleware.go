package auth

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
	m "github.com/t-bfame/diago/pkg/model"
	sto "github.com/t-bfame/diago/pkg/storage"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			// Allow unauthenticated users in
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			//validate jwt token
			tokenStr := header
			username, err := ParseToken(tokenStr)
			if err != nil {
				log.Error("Invalid token")
				http.Error(w, "Invalid token", http.StatusForbidden)
				return
			}

			// check if user exists in db
			foundUser, err := sto.GetUserByUserId(m.UserID(username))
			if err != nil {
				log.Error("Error getting user in authentication process")
				next.ServeHTTP(w, r)
				return
			}
			// hide password
			foundUser.Password = ""

			// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, foundUser)

			// and call the next handler with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func GetUserForContext(ctx context.Context) *m.User {
	user, ok := ctx.Value(userCtxKey).(*m.User)
	if !ok {
		log.Info("Retrieving context value failed")
		return nil
	}
	return user
}
