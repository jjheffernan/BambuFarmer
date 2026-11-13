module github.com/jjheffernan/BambuFarmer/services/edge

go 1.26.0

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/events v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/farm v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/objectstore v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/printerstate v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/minio/crc64nvme v1.0.2 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.95 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/tinylib/msgp v1.3.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api

replace github.com/jjheffernan/BambuFarmer/packages/printerstate => ../../packages/printerstate

replace github.com/jjheffernan/BambuFarmer/packages/events => ../../packages/events

replace github.com/jjheffernan/BambuFarmer/packages/farm => ../../packages/farm

replace github.com/jjheffernan/BambuFarmer/packages/objectstore => ../../packages/objectstore
