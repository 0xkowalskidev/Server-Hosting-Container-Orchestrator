Finish the api calls - ✓

Refactor api - ✓

etcd - ✓

Use container config - ✓

Move api wrapper to orchestrator - ✓

Nodes send status of containers - ✓

Cleanup api requests, task request struct from model - ✓

Main config - ✓

Bug, node deletes itself when deleting a container or restarting -  ✓
Follow up, delete containers from node when deleting a container - ✓

One namespace or many? - One
Switch to one namespace - ✓
Ensure namespace matches cfg when changing state -

Figure out how to watch status -  ✓
Follow up, seems there are waaaay to many events being triggered - 

Logs API call - ✓
Follow up, Performance concerns? -
Follow up, Where do we store the logs? - 

Stop hardcoding timeout -  

Switch delete to mark for deletion instead of just removing container from state -

Crash loop back off -

Make sure nodes send back usage/resources to control node - 

Make schedular schedule containers based on actual pertinent information -

Node currently overwrites node in etcd every time agent starts - 

Container delete api call always returns success - 

Container networking - 

Container storage -

Network cluster -

Auth/security -

Remove fatal errors, make sure agent cant crash -

Tests - 

Controllers/managers?
Consensus?
