package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	SessionKey	= "session"
)

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	app := appcontext.GetInstance()

	isToken := false
	var cookieValue string
	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("no kratos cookie or auth header")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		cookieValue = strings.Split(authHeader, " ")[1]
		isToken = strings.HasPrefix(cookieValue, "ory_st_")
	} else {
		cookieValue = cookie.Value
	}

	kratosSession, err := getKratosSession(cookieValue, isToken)
	if err != nil {
		var kratosError InvalidKratosSessionError
		if errors.As(err, &kratosError) {
			log.Printf("Invalid Kratos session cookie presented: %s", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		log.Printf("Problem verifying Kratos session: %s", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	a, err := kratosSession.CreateOrLoadAccount()
	if err != nil {
		log.Printf("Error retrieving account from Kratos session: %s", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	// Make a copy of user details and save into cookie - too much data violates cookie size rules
	userValues := *a
	if a.Profile != nil {
		userValues.Profile = nil
		profileCopy := *a.Profile
		profileCopy.Thumbnail = nil
		userValues.Profile = &profileCopy
	}
	session, _ := app.SESSION.Get(r, "session")
	valuesJson, _ := ffjson.Marshal(userValues)
	session.Values["user"] = valuesJson
	// saves session into response writer headers and returns response
	sessionError := app.SESSION.Save(r, w, session)
	if sessionError != nil {
		log.Printf("Error in session save: %s", sessionError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// Use original user details and return to app
	aJson, _ := ffjson.Marshal(a)
	log.Printf("userValues json %+v", userValues)
	// Return the profile information
	w.Header().Add("Content-Type", "application/json")
	w.Write(aJson)
}

func HandleAppVersion(w http.ResponseWriter, r *http.Request) {
	app := appcontext.GetInstance()

	// user's details
	contextUser := r.Context().Value(CurrentUserContextKey)
	if contextUser == nil {
		log.Printf("no user account in request context")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user := contextUser.(profile.Account)

	session, _ := app.SESSION.Get(r, "session")

	valuesJson, _ := ffjson.Marshal(user)
	session.Values["user"] = valuesJson

	sessionError := app.SESSION.Save(r, w, session)
	if sessionError != nil {
		log.Printf("Error in session save: %s", sessionError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
