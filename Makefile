REFLEX_REGEX='\.(go|html|css|js|nix)$$'
PAM_CFLAGS := $(shell pkg-config --cflags pam)
PAM_LDFLAGS := $(shell pkg-config --libs pam)

.PHONY: dev-worker-node
dev-worker-node:
	sudo -E env CGO_CFLAGS="$(PAM_CFLAGS)" CGO_LDFLAGS="$(PAM_LDFLAGS)" SFTP_KEY_PATH=/home/kowalski/.ssh/github_rsa SFTP_PORT=2022 LOGS_PATH=/home/kowalski/dev/server-hosting/container-orchestrator/logs MOUNTS_PATH=/home/kowalski/dev/server-hosting/container-orchestrator/mounts NODE_ID=node_1 CONTROL_NODE_URI=http://localhost:3001/api reflex -r $(REFLEX_REGEX) -s -- go run ./cmd/worker_node/main.go

.PHONY: dev-control-node
dev-control-node:
	 reflex -r $(REFLEX_REGEX) -s -- go run cmd/control_node/main.go

.PHONY: dev-control-panel
dev-control-panel:
	 reflex -r $(REFLEX_REGEX) -s -- go run cmd/control_panel/main.go

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

.PHONY: reset-logs
reset-logs:
	sudo rm -r logs/* 

.PHONY: reset-full
reset-full:
	make reset-ctr
	make reset-etcd
	make reset-network
	make reset-logs

.PHONY: cloc
cloc: 
	cloc --exclude-dir=.direnv --not-match-f='_test\.go$$' --by-file .

