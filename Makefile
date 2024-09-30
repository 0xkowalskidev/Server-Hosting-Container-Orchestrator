REFLEX_REGEX='\.go$|\.html$$'

.PHONY: dev-worker-node
dev-worker-node:
	sudo NODE_ID=node_1 CONTROL_NODE_URI=http://localhost:3000/api reflex -s -- go run ./cmd/worker_node/main.go

.PHONY: dev-control-node
dev-control-node:
	 reflex -s -- go run cmd/control_node/main.go

.PHONY: test
test:
	make test-worker-node


.PHONY: test-worker-node
test-worker-node:
	sudo go test -count=1 -v ./...

.PHONY: reset-ctr
reset-ctr:
	# Stop and delete all running tasks
	@sudo ctr -n $(CONTAINERD_NAMESPACE) tasks ls -q | xargs -r -I{} sudo ctr -n $(CONTAINERD_NAMESPACE) tasks kill -s SIGKILL {} || true
	# Ensure all tasks are stopped before deletion
	@sleep 1
	@sudo ctr -n $(CONTAINERD_NAMESPACE) tasks ls -q | xargs -r -I{} sudo ctr -n $(CONTAINERD_NAMESPACE) tasks delete {} || true
	# Delete all containers
	@sudo ctr -n $(CONTAINERD_NAMESPACE) containers ls -q | xargs -r -I{} sudo ctr -n $(CONTAINERD_NAMESPACE) containers delete {} || true

.PHONY: reset-etcd
reset-etcd:
	etcdctl del "" --prefix

.PHONY: reset-full
reset-full:
	make reset-ctr
	make reset-etcd
