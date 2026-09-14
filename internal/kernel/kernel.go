package kernel

import (
	"fmt"
	"net/http"

	"waypay_dedicated/internal/rbac"
	"waypay_dedicated/sdk"
)

type Kernel struct {
	mux         *http.ServeMux
	rbac        *rbac.Engine
	log         func(string, ...any)
	routes      []sdk.Route
	permissions map[string]sdk.Permission
}

func New(rbacEngine *rbac.Engine, log func(msg string, args ...any)) *Kernel {
	return &Kernel{
		rbac:        rbacEngine,
		log:         log,
		mux:         http.NewServeMux(),
		permissions: map[string]sdk.Permission{},
	}
}

func (k *Kernel) Register(m sdk.Module) error {
	manifest := m.Manifest()

	for _, p := range manifest.Permissions {
		if _, exists := k.permissions[p.Key]; exists {
			return fmt.Errorf("permission %q already registered", p.Key)
		}
		k.permissions[p.Key] = p
		k.log("permission registered", "key", p.Key, "module", manifest.Name)
	}

	for _, r := range manifest.Roles {
		if k.rbac.HasRole(r.Name) {
			return fmt.Errorf("role %q already registered", r.Name)
		}
		for _, key := range r.Permissions {
			if _, ok := k.permissions[key]; !ok {
				return fmt.Errorf("role %q references unknown permission %q", r.Name, key)
			}
		}
		k.rbac.AddRole(rbac.Role{Name: r.Name, Permissions: r.Permissions})
	}

	ctx := &sdk.Context{
		RegisterRoute: func(route sdk.Route) {
			k.routes = append(k.routes, route)
		},
		Log: k.log,
	}

	if err := m.Register(ctx); err != nil {
		return fmt.Errorf("module %s register: %w", manifest.Name, err)
	}

	k.log("module registered", "name", manifest.Name, "version", manifest.Version)
	return nil
}

func (k *Kernel) Routes() []sdk.Route { return k.routes }
func (k *Kernel) RBAC() *rbac.Engine  { return k.rbac }
