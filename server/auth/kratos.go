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

type InvalidKratosSessionError error

type VerifiableAddress struct {
	Value    string `json:"value"`
	Verified bool   `json:"verified"`
}

type KratosSession struct {
	Identity struct {
		ID     uuid.UUID `json:"id"`
		Traits struct {
			Email string `json:"email"`
		} `json:"traits"`
		VerifiableAddresses []VerifiableAddress    `json:"verifiable_addresses"`
		MetadataPublic      map[string]interface{} `json:"metadata_public"`
	} `json:"identity"`
}

var httpClient utils.HTTPClient

func init() {
	httpClient = &http.Client{}
}

func getKratosSession(cookieString string, isToken bool) (*KratosSession, error) {
	url, ok := os.LookupEnv("KRATOS_SESSION_URL")
	if !ok {
		return nil, errors.New("no KRATOS_SESSION_URL set in env")
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
		return nil, InvalidKratosSessionError(fmt.Errorf("kratos returned %d", resp.StatusCode))
	}

	var s KratosSession
	dec := ffjson.NewDecoder()
	err = dec.DecodeReader(resp.Body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *KratosSession) CreateOrLoadAccount() (*profile.Account, error) {
	ID := s.Identity.ID
	log.Printf("Identity in CreateOrLoad: %v, ID: %v", s.Identity, s.Identity.ID)
	a := new(profile.Account)
	a.ID = &ID
	a.ProfileID = &ID
	err := a.LoadByID()
	if err != nil {
		log.Printf("Error in CreateOrLoad LoadByID ACCOUNT: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					a = &profile.Account{
						UUIDBaseModel: utils.UUIDBaseModel{
							ID: &ID,
						},
						DateOfBirth:   nil,
						ProfileID:     nil,
						Email:         s.Identity.Traits.Email,
						EmailVerified: emailVerified(s.Identity.Traits.Email, s.Identity.VerifiableAddresses),
					}
					if err := a.Create(); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if s.Identity.Traits.Email != a.Email {
		a.Email = s.Identity.Traits.Email
		a.EmailVerified = emailVerified(s.Identity.Traits.Email, s.Identity.VerifiableAddresses)

		if err := a.UpdateEmail(); err != nil {
			log.Printf("Found new email in identity, error updating email in DB: %v", err)
			return nil, err
		}
	}

	roles, ok := s.Identity.MetadataPublic["roles"]
	if ok {
		for _, v := range roles.([]interface{}) {
			a.Roles = append(a.Roles, v.(string))
		}
	}

	return a, nil
}

func emailVerified(targetAddress string, addresses []VerifiableAddress) bool {
	for _, a := range addresses {
		if a.Value == targetAddress {
			return a.Verified
		}
	}
	return false
}
