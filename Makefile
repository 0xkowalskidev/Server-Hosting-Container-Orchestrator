.PHONY: dev-worker-node
dev-worker-node:
	sudo go run ./cmd/worker_node/main.go

.PHONY: reset-ctr
reset-ctr:
	# Stop and delete all running tasks
	@sudo ctr -n $(NAMESPACE_MAIN) tasks ls -q | xargs -r -I{} sudo ctr -n $(NAMESPACE_MAIN) tasks kill -s SIGKILL {} || true
	# Ensure all tasks are stopped before deletion
	@sleep 1
	@sudo ctr -n $(NAMESPACE_MAIN) tasks ls -q | xargs -r -I{} sudo ctr -n $(NAMESPACE_MAIN) tasks delete {} || true
	# Delete all containers
	@sudo ctr -n $(NAMESPACE_MAIN) containers ls -q | xargs -r -I{} sudo ctr -n $(NAMESPACE_MAIN) containers delete {} || true

.PHONY: test
test:
	sudo NAMESPACE_MAIN=test make reset-ctr
	sudo NAMESPACE_MAIN=test go test ./...
