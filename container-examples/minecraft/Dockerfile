# Use an official Java runtime as a parent image
FROM openjdk:17-slim

# Set the working directory in the container
WORKDIR /usr/src/minecraft

# Copy the Minecraft server jar file into the container
COPY server.jar /usr/src/minecraft

# Accept Minecraft's EULA
RUN echo "eula=true" > eula.txt

# Expose the default Minecraft port
EXPOSE 25565

# Run the Minecraft server
CMD ["java", "-Xmx1024M", "-Xms1024M", "-jar", "server.jar", "nogui"]

