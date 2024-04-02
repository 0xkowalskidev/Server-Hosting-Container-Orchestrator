#!/bin/sh

# Specify your namespace
NAMESPACE="example"

# List all containers in the specified namespace and delete them
for container in $(ctr -n $NAMESPACE containers list -q); do
    echo "Stopping and deleting container: $container"
    ctr -n $NAMESPACE task kill "$container"
    ctr -n $NAMESPACE task delete "$container"
    ctr -n $NAMESPACE containers delete "$container"
done

# Clean up snapshots, filtering out non-snapshot lines
ctr -n $NAMESPACE snapshots ls | tail -n +2 | awk '{print $1}' | while read -r snapshot; do
    echo "Deleting snapshot: $snapshot"
    ctr -n $NAMESPACE snapshots rm "$snapshot"
done

# Optionally, clean up images in the specified namespace
for image in $(ctr -n $NAMESPACE images list -q); do
    echo "Deleting image: $image"
    ctr -n $NAMESPACE images remove "$image"
done

echo "Cleanup complete."

