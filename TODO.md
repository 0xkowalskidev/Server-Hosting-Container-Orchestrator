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

## Switch delete to mark for deletion instead of just removing container from state -

## Container networking -  ✓

## Container storage - ✓
### Storage should be in its own package - ✓
### Delete storage on container delete - ✓
### Schedular should always cleanup unused storage -
### Schedular should bare storage in mind when scheduling -
### Node config should define how much storage they have -
### Node should use a different volume/partition just for container storage -
### Node should enforce storage limits - 

## Network cluster -

## Storage/networking should be handled separately from runtime in agent -

## Figure out how to containerize the dev environment (nix?) - 

# Full Refactor

# Later

## Logs & Monitoring (low level, not container level) -

## Create containerd model for containers -

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
