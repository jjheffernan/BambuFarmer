module github.com/jjheffernan/BambuFarmer/services/edge

go 1.26.0

require (
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/events v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/farm v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/printerstate v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.5.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api

replace github.com/jjheffernan/BambuFarmer/packages/printerstate => ../../packages/printerstate

replace github.com/jjheffernan/BambuFarmer/packages/events => ../../packages/events

replace github.com/jjheffernan/BambuFarmer/packages/farm => ../../packages/farm
