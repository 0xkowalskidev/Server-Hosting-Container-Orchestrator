.PHONY: dev-worker-node
dev-worker-node:
	sudo go run ./cmd/worker_node/main.go

CTR_NAMESPACE=default

.PHONY: reset-ctr
reset-ctr:
	# Stop and delete all running tasks
	@sudo ctr -n $(CTR_NAMESPACE) tasks ls -q | xargs -r -I{} sudo ctr -n $(CTR_NAMESPACE) tasks kill -s SIGKILL {} || true
	# Ensure all tasks are stopped before deletion
	@sleep 1
	@sudo ctr -n $(CTR_NAMESPACE) tasks ls -q | xargs -r -I{} sudo ctr -n $(CTR_NAMESPACE) tasks delete {} || true

	# Delete all containers
	@sudo ctr -n $(CTR_NAMESPACE) containers ls -q | xargs -r -I{} sudo ctr -n $(CTR_NAMESPACE) containers delete {} || true

.PHONY: test
test:
	sudo go test ./...
