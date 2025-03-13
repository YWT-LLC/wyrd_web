package appcontext

import (
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

type AppContext struct {
	DB         *gorm.DB
	SESSION		 sessions.Store
}

var GlobalContext *AppContext

func init() {
	GlobalContext = &AppContext{}
}

func GetInstance() *AppContext {
	return GlobalContext
}

func GetDB() *gorm.DB {
	return GetInstance().DB
}
