# MVP

## Finish the api calls - ✓

## Refactor api - ✓
### Container delete api call always returns success -

## etcd - ✓

## Use container config - ✓

## Move api wrapper to orchestrator - ✓

## Nodes send status of containers - ✓

## Cleanup api requests, task request struct from model - ✓

## Main config - ✓

## Bug, node deletes itself when deleting a container or restarting -  ✓
### Follow up, delete containers from node when deleting a container - ✓

## Node currently overwrites node in etcd every time agent starts - ✓

## One namespace or many? - One
### Switch to one namespace - ✓
### Ensure namespace matches cfg when changing state -

## Figure out how to watch status -  ✓
### Follow up, seems there are waaaay to many events being triggered - 

## Logs API call - ✓
### Follow up, Performance concerns? -
### Follow up, Where do we store the logs? -
### Follow up, Use htmx sse -
### Follow up, web crashes, proxy panics -

## Add resource constraints to containers - ✓

## Make sure nodes send back usage/resources to control node -
### Set node resource limits manually? - Yes!
### Could set up node resource automatically on initial join - 

## Make schedular schedule containers based on actual pertinent information - ✓

## Stop hardcoding timeout - ✓

## Container networking -  ✓
### Seperate into its own package - ✓
### Schedular should cleanup unused routes - ✓
### Schedular should consider ports when scheduling -  ✓

## Container storage - ✓
### Storage should be in its own package - ✓
### Delete storage on container delete - ✓
### Schedular should always cleanup unused storage - ✓
### Node should enforce storage limits -  ✓ 
### Node should define how much storage they have - ✓
### Schedular should bare storage in mind when scheduling - ✓


## Network cluster -
### Move worker/control plane into seperate CMD files - ✓
### Make worker node go through a phase of self discovery -
### Make control plane validate worker nodes ip before storing - 

## Live container metrics - 

## Set up build process -


# Full Refactor

## Network manager -

## Switch from resourceUsed to resourceAllocated on node as its misleading - 

## syncContainers - 

# Later

## Schedular should pick a port if none provided -

## Does not seem like cpu limit is working properly - 

## Schedular does not attempt to reschedule a node until a new node is created, might be okay but need to handle - 

## Storage sync wont cleanup bad img files - 

## Switch delete to mark for deletion instead of just removing container from state -

## Figure out how to containerize the dev environment (nix?) - 

## Node should use a different volume/partition just for container storage -

## Switch to SSE for worker node agent instead of polling - 

## Make sure node agent always syncs on reconnect -

## Logs & Monitoring (low level, not container level) -

## Create containerd model for containers? -

## Add taints -

## Tests - 

## Crash loop back off -

## Auth/security -

## Handle state race condition issues - 

## Remove fatal errors, make sure agent cant crash -

## sftp daemonset -

## Readiness probe -

## Controllers/managers?
## Consensus?

## Remove namespaces from routes as we are setting it in config anyway -

## Permissions -

## Config should define acceptable port range (e.g 30000-32767) -

## Node healthcheck should check if network is using any of the needed ports when it shouldent (e.g, 30000 is taken by a non container) - 

## lost+found will probably be put back on the volume if things crash -

## Storage sync could also check for non dirs and remove them aswell, at the moment it just guesses that they exist - 
