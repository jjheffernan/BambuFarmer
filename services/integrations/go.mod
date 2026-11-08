module github.com/jjheffernan/BambuFarmer/services/integrations

go 1.26.0

require (
	github.com/jjheffernan/BambuFarmer/packages/api v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/config v0.0.0
	github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0
)

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api

replace github.com/jjheffernan/BambuFarmer/packages/config => ../../packages/config
