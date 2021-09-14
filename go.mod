module github.com/wostzone/thingdir

go 1.14

require (
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.12
	github.com/kr/pretty v0.1.0 // indirect
	github.com/ohler55/ojg v1.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hubauth v0.0.0-00010101000000-000000000000
	github.com/wostzone/hubclient-go v0.0.0-00010101000000-000000000000
	github.com/wostzone/hubserve-go v0.0.0-20210907050346-343a1e9f8ad6
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Until stable
replace github.com/wostzone/hubclient-go => ../hubclient-go

replace github.com/wostzone/hubserve-go => ../hubserve-go

replace github.com/wostzone/hubauth => ../hubauth
