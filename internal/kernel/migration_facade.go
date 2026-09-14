package kernel

import (
	"waypay_dedicated/internal/migrator"
	"waypay_dedicated/sdk"
)

type migrationFacade struct {
	migrator *migrator.Migrator
	module   string
}

func (f *migrationFacade) Register(migs ...sdk.Migration) error {
	internal := make([]migrator.Migration, 0, len(migs))
	for _, m := range migs {
		internal = append(internal, migrator.Migration{
			Version: m.Version,
			Name:    m.Name,
			UpSQL:   m.UpSQL,
			DownSQL: m.DownSQL,
		})
	}
	return f.migrator.Register(f.module, internal...)
}
