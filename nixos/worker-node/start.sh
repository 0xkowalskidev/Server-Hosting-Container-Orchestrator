QEMU_KERNEL_PARAMS="console=ttyS0" \
QEMU_NET_OPTS="hostfwd=tcp::8081-:8081,hostfwd=tcp::2223-:22,hostfwd=tcp::30001-:30001" \
./result/bin/run-worker-node-vm -nographic -m 16G -smp 4;

