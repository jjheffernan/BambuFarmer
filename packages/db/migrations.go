package db

// Schema is defined in Go (not committed .sql files). Names keep a .sql suffix so
// existing schema_migrations rows from earlier embeds continue to match.

type migration struct {
	Name string
	SQL  string
}

// orderedMigrations is applied by Migrate in declaration order.
var orderedMigrations = []migration{
	{Name: "001_baseline.sql", SQL: migration001Baseline},
	{Name: "002_users.sql", SQL: migration002Users},
	{Name: "003_job_file.sql", SQL: migration003JobFile},
}
