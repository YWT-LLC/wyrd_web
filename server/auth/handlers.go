package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"
	
	"github.com/pquerna/ffjson/ffjson"
)

const (
	SessionKey = "session"
	UserKey		 = "user"
)

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	app := appcontext.GetInstance()
	isToken := false
	
	// Retrieve Kratos session cookie or authorization header
	cookie, err := r.Cookie("ory_kratos_session")
	var cookieValue string

	if err != nil {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("No Kratos cookie or auth header")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		cookieValue = strings.Split(authHeader, " ")[1]
		isToken = strings.HasPrefix(cookieValue, "ory_st_")
	} else {
		cookieValue = cookie.Value
	}
		
	// Verify Kratos session
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
	
	// Retrieve account from Kratos session
	account, err := kratosSession.CreateOrLoadAccount()
	if err != nil {
		log.Printf("Error retrieving account from Kratos session: %s", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	
	// Prepare user values for session
	userValues := *account
	if account.Profile != nil {
		userValues.Profile = nil
		pCopy := *account.Profile
		pCopy.Thumbnail = nil
		userValues.Profile = &pCopy
	}
	
	// Save user values in session
	session, _ := app.SESSION.Get(r, SessionKey)
	valuesJson, _ := ffjson.Marshal(userValues)
	session.Values[UserKey] = valuesJson
	
	if err := app.SESSION.Save(r, w, session); err != nil {
		log.Printf("Error in session save: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	
	// Respond with user account JSON
	accountJson, _ := ffjson.Marshal(account)
	log.Printf("User values JSON: %+v", userValues)
	w.Header().Add("Content-Type", "application/json")
	w.Write(accountJson)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	app := appcontext.GetInstance()
	
	// Retrieve user from context
	contextUser := r.Context().Value(CurrentUserContextKey)
	if contextUser == nil {
		log.Printf("No user account in request context")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	
	user := contextUser.(profile.Account)
	
	// Save user values in session
	session, _ := app.SESSION.Get(r, SessionKey)
	valuesJson, _ := ffjson.Marshal(user)
	session.Values[UserKey] = valuesJson
	
	if err := app.SESSION.Save(r, w, session); err != nil {
		log.Printf("Error in session save: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}