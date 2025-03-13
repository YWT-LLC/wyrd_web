package auth

import (
	"context"
	"log"
	"net/http"
	
	"github.com/gorilla/mux"
	"github.com/pquerna/ffjson/ffjson"
)

type AuthKey string

const (
	SessionKey	= "session"
	SessionTail	= "/session"
	SessionPath	= "/auth/session"
	UserKey			= AuthKey("currentUser")
)

func AddRoutes(router *mux.Router) {
	router.Path(SessionTail).Methods(http.MethodGet).Handler(http.HandlerFunc(sessionHandler))
	router.Path(SessionTail).Methods(http.MethodPut).Handler(ValidateSession(http.HandlerFunc(versionHandler)))
}

func ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := appcontext.GetInstance()
		
		session, _ := app.SESSION.Get(r, SessionKey)
		if session.IsNew {
			log.Println("No session found")
			http.Redirect(w, r, SessionPath, http.StatusSeeOther)
			return
		}
		
		currentUserJson, ok := session.Values[UserSessionKey]
		if !ok {
			log.Println("The session does not contain a user")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		
		var currentUser profile.Account
		if err := ffjson.Unmarshal(currentUserJson.([]byte), &currentUser); err != nil {
			log.Println("Unable to Unmarshal the user info:", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		
		ctx := context.WithValue(r.Context(), CurrentUserKey, currentUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}