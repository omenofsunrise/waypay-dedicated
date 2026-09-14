package sdk

import (
	"database/sql"
	"net/http"
)

type Module interface {
	Manifest() Manifest
	Register(ctx *Context) error
}

type Manifest struct {
	Name        string
	Version     string
	Permissions []Permission
	Roles       []Role
}

type Permission struct {
	Key         string
	Description string
}

type Role struct {
	Name        string
	Permissions []string
}

type Context struct {
	RegisterRoute func(Route)
	Log           func(msg string, args ...any)
	DB            *sql.DB // all modules are trusted for now (Not AI generated)
	Migrations    MigrationRunner
}

type Route struct {
	Method       string
	Path         string
	RequiredPerm string
	Handler      http.HandlerFunc
}
