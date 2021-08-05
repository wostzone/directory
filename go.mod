module github.com/wostzone/thingdir-go

go 1.14

require (
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.12
	github.com/kr/pretty v0.1.0 // indirect
	github.com/ohler55/ojg v1.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/wostlib-go v0.0.0-20210804043428-0505ddce2f64
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Until stable
replace github.com/wostzone/wostlib-go => ../wostlib-go
