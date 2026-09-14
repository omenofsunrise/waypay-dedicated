package migrator

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"sync"
)

type Migration struct {
	Module  string
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

type Migrator struct {
	db      *sql.DB
	mu      sync.Mutex
	pending map[string][]Migration
}

func New(db *sql.DB) *Migrator {
	return &Migrator{
		db:      db,
		pending: map[string][]Migration{},
	}
}

func (m *Migrator) Register(moduleName string, migs ...Migration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, mig := range migs {
		if mig.Version <= 0 {
			return fmt.Errorf("migration version must be > 0 (module %s, name %s)",
				moduleName, mig.Name)
		}
		mig.Module = moduleName
		m.pending[moduleName] = append(m.pending[moduleName], mig)
	}
	return nil
}

func (m *Migrator) ApplyAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.ensureMigrationsTable(); err != nil {
		return err
	}

	modules := make([]string, 0, len(m.pending))
	for mod := range m.pending {
		modules = append(modules, mod)
	}
	sort.Strings(modules)

	for _, mod := range modules {
		migs := m.pending[mod]
		sort.Slice(migs, func(i, j int) bool {
			return migs[i].Version < migs[j].Version
		})

		applied, err := m.appliedVersions(mod)
		if err != nil {
			return err
		}

		for _, mig := range migs {
			if applied[mig.Version] {
				continue
			}
			if err := m.applyOne(mig); err != nil {
				return fmt.Errorf("apply migration %s/%d (%s): %w",
					mig.Module, mig.Version, mig.Name, err)
			}
			log.Printf("migration applied: %s/%d %s", mig.Module, mig.Version, mig.Name)
		}
	}
	return nil
}

func (m *Migrator) ensureMigrationsTable() error {
	_, err := m.db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            module     TEXT NOT NULL,
            version    INT  NOT NULL,
            name       TEXT NOT NULL,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
            PRIMARY KEY (module, version)
        )
    `)
	return err
}

func (m *Migrator) appliedVersions(module string) (map[int]bool, error) {
	rows, err := m.db.Query(
		`SELECT version FROM schema_migrations WHERE module = $1`, module)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func (m *Migrator) applyOne(mig Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(mig.UpSQL); err != nil {
		return fmt.Errorf("up sql: %w", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (module, version, name) VALUES ($1, $2, $3)`,
		mig.Module, mig.Version, mig.Name,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}
