package sdk

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

type MigrationRunner interface {
	Register(migrations ...Migration) error
}
