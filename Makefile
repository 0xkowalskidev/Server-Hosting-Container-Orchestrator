.PHONY: dev-control-node
dev-control-node:
	echo "Starting control-node..."
	go run ./cmd/control-node/control-node.go

.PHONY: dev-worker-node
dev-worker-node:
	echo "Starting worker-node..."
	go run ./cmd/worker-node/worker-node.go
