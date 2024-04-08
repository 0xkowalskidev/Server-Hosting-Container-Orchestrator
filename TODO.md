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

## Make schedular schedule containers based on actual pertinent information - ✓

## Stop hardcoding timeout - ✓

## Switch delete to mark for deletion instead of just removing container from state -

## Container networking -  ✓
### Followup, Make more resiliant, should be done by agent on node pull, not on containter creation - 

## Container storage -

## Network cluster -

# Full Refactor

# Later

## Add taints -

## Tests - 

## Crash loop back off -

## Auth/security -

## Handle state race condition issues - 

## Remove fatal errors, make sure agent cant crash -

## sftp daemonset -

## Controllers/managers?
## Consensus?
