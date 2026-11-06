module github.com/jjheffernan/BambuFarmer/services/gateway

go 1.23.0

require (
	github.com/jjheffernan/BambuFarmer/packages/events v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0 // indirect
	github.com/redis/go-redis/v9 v9.7.3 // indirect
)

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api

replace github.com/jjheffernan/BambuFarmer/packages/events => ../../packages/events
