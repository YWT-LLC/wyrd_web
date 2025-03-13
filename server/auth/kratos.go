package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/ffjson/ffjson"
	"gorm.io/gorm"
)

type InvalidSession error

type EmailStatus struct {
	Value    string `json:"value"`
	Verified bool   `json:"verified"`
}

type KratosSession struct {
	Identity struct {
		ID     uuid.UUID `json:"id"`
		Traits struct {
			Name 	string 	`json:"name"`
			Email *string `json:"email"`
		} `json:"traits"`
		EmailStatuses 			[]EmailStatus    				`json:"email_statuses"`
		PublicMetadata      map[string]interface{}	`json:"public_metadata"`
	} `json:"identity"`
}

var httpClient utils.HTTPClient

func init() {
	httpClient = &http.Client{}
}

func getKratosSession(cookieString string, isToken bool) (*KratosSession, error) {
	url, ok := os.LookupEnv("KRATOS_SESSION_URL")
	if !ok {
		return nil, errors.New("No KRATOS_SESSION_URL in env")
	}

	client := httpClient
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	expiry := time.Now().AddDate(1, 0, 0)

	if isToken {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cookieString))
	} else {
		cookie := http.Cookie{Value: cookieString, Name: "ory_kratos_session", Path: "/", Expires: expiry, MaxAge: 31536000}
		req.AddCookie(&cookie)
	}

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, InvalidSession(fmt.Errorf("Kratos returned %d", resp.StatusCode))
	}

	var session KratosSession
	dec := ffjson.NewDecoder()
	err = dec.DecodeReader(resp.Body, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (session *KratosSession) FetchAccount() (*profile.Account, error) {
	ID := session.Identity.ID
	log.Printf("Fetched Identity: %v, ID: %v", session.Identity, session.Identity.ID)

	account := new(profile.Account)
	account.ID = &ID
	account.ProfileID = &ID

	if err := account.LoadByID(); err != nil {
		log.Printf("LoadByID error: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					account = &profile.Account{
						UUIDBaseModel: utils.UUIDBaseModel{
							ID: &ID,
						},
						ProfileID:     nil,
						Email:         session.Identity.Traits.Email,
						EmailVerified: getEmailStatus(session.Identity.Traits.Email, session.Identity.EmailStatuses),
					}
					if err := account.Create(); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if session.Identity.Traits.Email != account.Email {
		account.Email = session.Identity.Traits.Email
		account.EmailVerified = getEmailStatus(session.Identity.Traits.Email, session.Identity.EmailStatuses)

		if err := account.UpdateEmail(); err != nil {
			log.Printf("Found new email in identity, error updating email in DB: %v", err)
			return nil, err
		}
	}

	roles, ok := session.Identity.MetadataPublic["roles"]
	if ok {
		for _, v := range roles.([]interface{}) {
			account.Roles = append(account.Roles, v.(string))
		}
	}

	return account, nil
}

func getEmailStatus(address string, statuses []EmailStatus) bool {
	for _, account := range statuses {
		if account.Value == address {
			return account.Verified
		}
	}
	return false
}
