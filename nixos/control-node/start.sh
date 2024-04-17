QEMU_KERNEL_PARAMS="console=ttyS0" QEMU_NET_OPTS="hostfwd=tcp::8080-:8080,hostfwd=tcp::2222-:22" ./result/bin/run-control-node-vm -nographic 
