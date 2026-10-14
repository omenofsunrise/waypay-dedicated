package persons

import (
	"fmt"
	"net/http"
	"waypay_dedicated/sdk"
)

const (
	PermRead   = "persons.read"
	PermCreate = "persons.create"
	PermUpdate = "persons.update"
	PermDelete = "persons.delete"
)

type Module struct {
	repo *Repo
	log  func(msg string, args ...any)
}

func New() sdk.Module { return &Module{} }

func (m *Module) Manifest() sdk.Manifest {
	return sdk.Manifest{
		Name:    "persons",
		Version: "1.0.0",
		Permissions: []sdk.Permission{
			{Key: PermRead, Description: "Просмотр лиц"},
			{Key: PermCreate, Description: "Создание лиц"},
			{Key: PermUpdate, Description: "Изменение лиц"},
			{Key: PermDelete, Description: "Мягкое удаление лиц"},
		},
		Roles: []sdk.Role{
			{
				Name:        "persons_admin",
				Permissions: []string{PermRead, PermCreate, PermUpdate, PermDelete},
			},
			{
				Name:        "persons_operator",
				Permissions: []string{PermRead, PermCreate, PermUpdate},
			},
			{
				Name:        "persons_viewer",
				Permissions: []string{PermRead},
			},
		},
	}
}

func (m *Module) Register(ctx *sdk.Context) error {
	if ctx.DB == nil {
		return fmt.Errorf("persons: ctx.DB is nil")
	}
	m.repo = NewRepo(ctx.DB)
	m.log = ctx.Log

	if err := ctx.Migrations.Register(migrations()...); err != nil {
		return fmt.Errorf("persons: register migrations: %w", err)
	}

	routes := []sdk.Route{
		{Method: http.MethodPost, Path: "/persons", RequiredPerm: PermCreate, Handler: m.handleCreate},
		{Method: http.MethodGet, Path: "/persons", RequiredPerm: PermRead, Handler: m.handleList},
		{Method: http.MethodGet, Path: "/persons/{id}", RequiredPerm: PermRead, Handler: m.handleGet},
		{Method: http.MethodPatch, Path: "/persons/{id}", RequiredPerm: PermUpdate, Handler: m.handleUpdate},
		{Method: http.MethodDelete, Path: "/persons/{id}", RequiredPerm: PermDelete, Handler: m.handleDelete},
	}
	for _, rt := range routes {
		ctx.RegisterRoute(rt)
	}
	return nil
}
