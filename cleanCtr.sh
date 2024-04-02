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


echo "Cleanup complete."

