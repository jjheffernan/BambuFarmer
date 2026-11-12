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
	{Name: "004_job_queue.sql", SQL: migration004JobQueue},
	{Name: "005_job_spool.sql", SQL: migration005JobSpool},
}
