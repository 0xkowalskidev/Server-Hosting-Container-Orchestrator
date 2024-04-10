#!/bin/sh

# Specify your namespace
NAMESPACE="development"

# List all containers in the specified namespace and delete them
for container in $(ctr -n $NAMESPACE containers list -q); do
    echo "Stopping and deleting container: $container"
    # Ensure the task is stopped before deletion
    if ctr -n $NAMESPACE task list | grep -q "$container"; then
        ctr -n $NAMESPACE task kill "$container" -s SIGKILL v0
        sleep 1 # Give a moment for the task to stop
        ctr -n $NAMESPACE task delete "$container"
    fi
    ctr -n $NAMESPACE containers delete "$container"
done

# Optionally, clean up snapshots
for snapshot in $(ctr -n $NAMESPACE snapshots ls | tail -n +2 | awk '{print $1}'); do
    echo "Deleting snapshot: $snapshot"
    ctr -n $NAMESPACE snapshots rm "$snapshot"
done

for image in $(ctr -n $NAMESPACE images list | tail -n +2 | awk '{print $1}'); do
    echo "Deleting image: $image"
    ctr -n $NAMESPACE images rm "$image"
done

rm  development-minecraft-server-1.log
rm  development-minecraft-server-2.log
sudo rm -rf ./mounts/minecraft-server-1


echo "Cleanup complete."

