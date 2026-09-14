package clients

import (
	"encoding/json"
	"net/http"
	"waypay_dedicated/sdk"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Manifest() sdk.Manifest {
	return sdk.Manifest{
		Name:    "clients",
		Version: "1.0.0",
		Permissions: []sdk.Permission{
			{Key: "clients.read", Description: "Read clients"},
			{Key: "clients.write", Description: "Create/update clients"},
		},
		Roles: []sdk.Role{
			{
				Name:        "user",
				Permissions: []string{"clients.read"},
			},
			{
				Name:        "admin",
				Permissions: []string{"clients.read", "clients.write"},
			},
		},
	}
}

func (m *Module) Register(ctx *sdk.Context) error {
	ctx.RegisterRoute(sdk.Route{
		Method:       "GET",
		Path:         "/api/clients",
		RequiredPerm: "clients.read",
		Handler:      listClients,
	})
	ctx.RegisterRoute(sdk.Route{
		Method:       "POST",
		Path:         "/api/clients",
		RequiredPerm: "clients.write",
		Handler:      createClient,
	})
	return nil
}

func listClients(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]string{
		{"id": "1", "name": "Acme"},
		{"id": "2", "name": "Globex"},
	})
}

func createClient(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}
