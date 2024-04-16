#!/bin/bash

# Check if /data/server is empty and copy files from /data/files if it is
if [ -d "/data/server" ] && [ -z "$(ls -A /data/server)" ]; then
    echo "/data/server is empty. Copying files from /data/files..."
    cp -r /data/files/* /data/server/
    echo "Files copied."
else
    echo "/data/server is not empty or does not exist."
fi

# Convert the JAVA_MEM environment variable from GB to MB
MEMORY_MB=$((MEMORY * 1024))

# Run Java with the calculated memory settings
echo "Starting Minecraft server with ${MEMORY_MB} MB of memory."
java -Xmx${MEMORY_MB}M -Xms${MEMORY_MB}M -jar /data/server/server.jar nogui

