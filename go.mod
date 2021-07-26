module github.com/wostzone/wostdir

go 1.14

require (
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/idprov-go v0.0.0-20210713173130-29613537efb2
	github.com/wostzone/wostlib-go v0.0.0-20210720190756-58073526660e
)

// Until stable
replace github.com/wostzone/wostlib-go => ../wostlib-go
