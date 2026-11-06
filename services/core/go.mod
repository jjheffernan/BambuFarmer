module github.com/jjheffernan/BambuFarmer/services/core

go 1.23.0

require (
	github.com/jackc/pgx/v5 v5.7.6
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/config v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/db v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/events v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/version v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/redis/go-redis/v9 v9.18.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api

replace github.com/jjheffernan/BambuFarmer/packages/config => ../../packages/config

replace github.com/jjheffernan/BambuFarmer/packages/db => ../../packages/db

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/events => ../../packages/events

replace github.com/jjheffernan/BambuFarmer/packages/version => ../../packages/version
