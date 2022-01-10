module github.com/scraimer/rpi0-ble-relay

go 1.17

require github.com/paypal/gatt v0.0.0-20151011220935-4ae819d591cf

require (
	github.com/deepmap/oapi-codegen v1.8.2 // indirect
	github.com/influxdata/influxdb-client-go/v2 v2.6.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/paypal/gatt => ./local/gatt
