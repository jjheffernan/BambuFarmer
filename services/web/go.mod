module github.com/jjheffernan/BambuFarmer/services/web

go 1.26.0

require github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0

require github.com/jjheffernan/BambuFarmer/packages/api v0.0.0 // indirect

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api
