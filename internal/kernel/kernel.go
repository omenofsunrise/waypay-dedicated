package kernel

import (
	"database/sql"
	"fmt"
	"net/http"

	"waypay_dedicated/internal/migrator"
	"waypay_dedicated/internal/rbac"
	"waypay_dedicated/sdk"
)

type Kernel struct {
	mux         *http.ServeMux
	rbac        *rbac.Engine
	log         func(string, ...any)
	routes      []sdk.Route
	permissions map[string]sdk.Permission
	db          *sql.DB
	migrator    *migrator.Migrator
}

func New(
	rbacEngine *rbac.Engine,
	mig *migrator.Migrator,
	db *sql.DB,
	log func(msg string, args ...any),
) *Kernel {
	return &Kernel{
		rbac:        rbacEngine,
		migrator:    mig,
		db:          db,
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
		DB:  k.db,
		Migrations: &migrationFacade{
			migrator: k.migrator,
			module:   manifest.Name,
		},
	}

	if err := m.Register(ctx); err != nil {
		return fmt.Errorf("module %s register: %w", manifest.Name, err)
	}

	k.log("module registered", "name", manifest.Name, "version", manifest.Version)
	return nil
}

func (k *Kernel) BuildMux() *http.ServeMux {
	mux := http.NewServeMux()
	for _, r := range k.routes {
		handler := k.withAuth(k.requirePerm(r.RequiredPerm, r.Handler))
		mux.HandleFunc(r.Method+" "+r.Path, handler)
	}
	return mux
}

func (k *Kernel) LoadAll(modules []sdk.Module) error {
	// TODO
	// 1) сбор манифестов
	// 2) проверка конфликтов и версий
	// 3) топологическая сортировка
	// 4) регистрация в правильном порядке

	loaded := map[string]bool{}
	for _, m := range modules {
		manifest := m.Manifest()
		if loaded[manifest.Name] {
			return fmt.Errorf("module %q loaded twice", manifest.Name)
		}
		if err := k.Register(m); err != nil {
			return fmt.Errorf("load module %s: %w", manifest.Name, err)
		}
		loaded[manifest.Name] = true
	}
	return nil
}

func (k *Kernel) Routes() []sdk.Route { return k.routes }
func (k *Kernel) RBAC() *rbac.Engine  { return k.rbac }
