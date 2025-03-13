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
	AdminKey		= "admin"
	UserKey			= AuthKey("currentUser")
)

func AddRoutes(router *mux.Router) {
	router.Path("/session").Methods(http.MethodGet).Handler(http.HandlerFunc(sessionHandler))
	router.Path("/session").Methods(http.MethodPut).Handler(ValidateSession(http.HandlerFunc(versionHandler)))
}

func ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := appcontext.GetInstance()
		
		session, _ := app.SESSION.Get(r, SessionKey)
		if session.IsNew {
			log.Println("unable to find existing session")
			http.Redirect(w, r, "/auth/session", http.StatusSeeOther)
			return
		}
		
		currentUserJson, ok := session.Values[UserSessionKey]
		if !ok {
			log.Println("error session doesn't contain user information")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		
		var currentUser profile.Account
		if err := ffjson.Unmarshal(currentUserJson.([]byte), &currentUser); err != nil {
			log.Println("error unmarshaling user information:", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		
		ctx := context.WithValue(r.Context(), CurrentUserKey, currentUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ValidateAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(CurrentUserKey).(profile.Account)
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		
		for _, role := range currentUser.Roles {
			if role == AdminKey {
				next.ServeHTTP(w, r)
				return
			}
		}
		
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	})
}