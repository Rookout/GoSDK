package rookout

//go:generate protoc --go_opt=paths=source_relative --go_out=./pkg/protobuf -I ./interfaces/proto -I ./interfaces/proto/include ./interfaces/proto/variant.proto ./interfaces/proto/variant2.proto ./interfaces/proto/envelope.proto ./interfaces/proto/messages.proto ./interfaces/proto/agent_info.proto ./interfaces/proto/controller_info.proto
