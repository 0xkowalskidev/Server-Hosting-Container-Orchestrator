#!/bin/bash

# Print the current working directory
echo "Current Working Directory:"
pwd

# Function to list contents of a directory if it exists
list_directory_contents() {
    dir=$1  # The directory to list
    echo "Contents of $dir:"
    if [ -d "$dir" ]; then
        ls "$dir"
    else
        echo "Directory does not exist."
    fi
    echo ""  # Print a newline for better readability
}

# List the contents of /data, /data/server, and /data/scripts
list_directory_contents "/data"
list_directory_contents "/data/server"
list_directory_contents "/data/scripts"
list_directory_contents "/data/files"

# Check if /data/server is empty and copy files from /data/files if it is
if [ -d "/data/server" ] && [ -z "$(ls -A /data/server)" ]; then
    echo "/data/server is empty. Copying files from /data/files..."
    cp -r /data/files/* /data/server/
    echo "Files copied."
else
    echo "/data/server is not empty or does not exist."
fi

java -Xmx1024M -Xms1024M -jar /data/server/server.jar nogui
