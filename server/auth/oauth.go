package main

import (
	"log"
	"net/http"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/gin-gonic/gin"
)

var oauthServer *server.Server

func initOAuth2() {
	manager := manage.NewDefaultManager()

	// Token storage
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// Client store
	clientStore := store.NewClientStore()
	clientStore.Set("app-client", &store.Client{
		ID:     "app-client",
		Secret: "app-secret",
		Domain: "http://localhost",
	})
	manager.MapClientStorage(clientStore)

	oauthServer = server.NewServer(server.NewConfig(), manager)
}

func TokenHandler(c *gin.Context) {
	err := oauthServer.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
