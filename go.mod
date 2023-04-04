module github.com/parca-dev/parca-push

go 1.19

require (
	github.com/alecthomas/kong v0.7.1
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/oklog/run v1.1.0
	github.com/parca-dev/parca v0.15.0
	github.com/prometheus/client_golang v1.14.0
	google.golang.org/grpc v1.53.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.14.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

// required by https://github.com/grpc-ecosystem/grpc-gateway/releases/tag/v2.10.3
replace cloud.google.com/go/storage v1.19.0 => cloud.google.com/go/storage v1.10.0
