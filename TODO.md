# Features
## Container status api call - Done
## Container metrics api should return cpu usage as a useful value - Done
## SFTP - Done
## SFTP perms currently insecure, user can access any dir! -
## Should handle sftp password - 

# Refactors
## Container status -
## Move logs/metrics api into api wrapper - 
## Move api wrapper into models or own dir -
## Containerd should ensure memory/storage limits on create

# Bugs
## Control Panel crash on server restart due to sse conn - Done
## Worker node does not properly create/remove portmappings if netns already exists, general weirdness if a container already exists on restart- 
## Container status race condition - 
## SFTP might not let user disconnect -
