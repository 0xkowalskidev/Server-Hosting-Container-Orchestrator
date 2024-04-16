.PHONY: dev-control-node
dev-control-node:
	echo "Starting control-node..."
	go run ./cmd/control-node/control-node.go

.PHONY: dev-worker-node
dev-worker-node:
	echo "Starting worker-node..."
	go run ./cmd/worker-node/worker-node.go

# You probably shouldent ever use this
.PHONY: reset-network
reset-network:
	sudo find /var/lib/cni/networks/mynet/ -name "10.22.0.*" -exec rm {} +
	sudo iptables -F
	sudo iptables -X
	sudo iptables -t nat -F
	sudo iptables -t nat -X
	sudo iptables -t mangle -F
	sudo iptables -t mangle -X
	sudo iptables -P INPUT ACCEPT
	sudo iptables -P FORWARD ACCEPT
	sudo iptables -P OUTPUT ACCEPT
	sudo umount /var/run/netns/*
	sudo rm /var/run/netns/*

.PHONY: reset-etcd
reset-etcd:
	etcdctl del "" --prefix

.PHONY: reset-ctr
reset-ctr:
	@NAMESPACE="development"; \
	# List all containers in the specified namespace and delete them \
	for container in $$(ctr -n $$NAMESPACE containers list -q); do \
    	echo "Stopping and deleting container: $$container"; \
    	# Ensure the task is stopped before deletion \
    	if ctr -n $$NAMESPACE task list | grep -q "$$container"; then \
        	ctr -n $$NAMESPACE task kill "$$container" -s SIGKILL v0; \
        	sleep 1; # Give a moment for the task to stop \
        	ctr -n $$NAMESPACE task delete "$$container"; \
    	fi; \
    	ctr -n $$NAMESPACE containers delete "$$container"; \
	done; \
	# Optionally, clean up snapshots \
	for snapshot in $$(ctr -n $$NAMESPACE snapshots ls | tail -n +2 | awk '{print $$1}'); do \
    	echo "Deleting snapshot: $$snapshot"; \
    	ctr -n $$NAMESPACE snapshots rm "$$snapshot"; \
	done; \
	for image in $$(ctr -n $$NAMESPACE images list | tail -n +2 | awk '{print $$1}'); do \
    	echo "Deleting image: $$image"; \
    	ctr -n $$NAMESPACE images rm "$$image"; \
	done; \
	echo "Cleanup complete. (except logs and mounts)"

.PHONY: reset-full
reset-full:
	make reset-etcd
	make reset-ctr
	make reset-network

