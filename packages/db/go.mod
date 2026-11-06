module github.com/jjheffernan/BambuFarmer/packages/db

go 1.23.0

require (
	github.com/jackc/pgx/v5 v5.7.4
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0
)

replace github.com/jjheffernan/BambuFarmer/packages/api => ../api
