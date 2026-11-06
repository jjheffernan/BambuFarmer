module github.com/jjheffernan/BambuFarmer/services/web

go 1.23.0

require github.com/jjheffernan/BambuFarmer/packages/httpserver v0.0.0

replace github.com/jjheffernan/BambuFarmer/packages/httpserver => ../../packages/httpserver

replace github.com/jjheffernan/BambuFarmer/packages/api => ../../packages/api
