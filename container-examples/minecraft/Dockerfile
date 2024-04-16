# Use an official Java runtime as a parent image
FROM openjdk:17-slim

# Copy the files and scripts to the container
COPY files /data/files
COPY scripts /data/scripts

# Set the working directory in the container
WORKDIR /data/server

# Expose the default Minecraft port
EXPOSE 25565

# Set ENV variables
# Memory, in GB, default is 1GB
ENV MEMORY=1

# Run the Minecraft server
CMD ["sh", "/data/scripts/start.sh"]

