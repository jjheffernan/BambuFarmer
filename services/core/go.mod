module github.com/jjheffernan/BambuFarmer/services/core

go 1.23.0

require (
	github.com/jackc/pgx/v5 v5.7.4
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/db v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0
)

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api

replace github.com/jjheffernan/BambuFarmer/packages/db => ../../packages/db

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver
