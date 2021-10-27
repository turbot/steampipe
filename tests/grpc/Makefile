all:
	go build -o kv
	go build -o kv-go-grpc ./plugin-go-grpc
	export KV_PLUGIN="./kv-go-grpc"