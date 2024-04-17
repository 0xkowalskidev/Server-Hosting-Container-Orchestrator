if [ -f ./worker-node.qcow2 ]; then
    # Resize the new disk image to 10GB
    qemu-img resize worker-node.qcow2 10G

    echo "Disk image created and resized to 10GB successfully. Run 'sudo resize2fs /dev/vda' on the vm"
else
    echo "Failed to resize the disk image."
fi


