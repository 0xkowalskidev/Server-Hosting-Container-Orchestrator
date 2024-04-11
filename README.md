# A "simple" container orchestrator

## Control Node Components
1. Start API

Function: Serves as the primary interface for users and external systems to interact with the orchestrator. It handles requests to create, update, delete, and manage container workloads and resources.

Implementation: Could be built using a high-performance HTTP framework like Gin for Go. The API would define endpoints for various management tasks and translate these requests into internal commands for the scheduler, controllers, and state manager.

2. Start Scheduler

Function: Responsible for deciding where to run new containers based on resource availability, constraints, policies, and workload requirements.

Implementation: The scheduler would need to be aware of the current state of all worker nodes and their resources. When it receives a request to schedule a container, it uses this information along with the container's requirements to select an appropriate node.

3. Start Controllers/Managers

Function: Maintain the desired state of the system. This includes controllers for handling node lifecycle, managing networking policies, ensuring the desired number of container replicas, and more.

Implementation: Each controller continuously watches the system's state and works to reconcile the current state with the desired state. They interact with the API and state manager to get updates and make changes.

4. Start State Manager

Function: Manages the persistent state of the cluster, including node information, configurations, and container states.

Implementation: Could be implemented using a distributed key-value store like etcd, which provides consistency and high availability of the cluster state.


## Worker Node Components

1. Start Runtime

Function: Manages the lifecycle of containers on the node, handling tasks like starting, stopping, and monitoring containers.

Implementation: This involves interfacing with a container runtime, such as containerd or CRI-O, which provides the low-level mechanisms for running containers.

2. Start Agent

Function: Communicates with the control plane, executes tasks assigned by the scheduler, reports the status of containers and resources, and ensures that its node's state matches what the control plane expects.

Implementation: The agent would be a daemon running on each worker node, constantly syncing with the control node's API to receive commands and send updates.

3. Start Networking

Function: Handles the networking aspects for containers, ensuring they can communicate with each other and with the outside world according to defined policies.

Implementation: Could leverage existing networking solutions like CNI plugins, which allow for a flexible and extensible networking setup that can be tailored to different environments and requirements.

4. Start Local Storage

Function: Manages local storage resources for containers, providing and managing volumes for persistent data storage.

Implementation: This component would interface with the underlying storage system to provision, mount, and manage storage volumes for containers, potentially integrating with cloud storage APIs or local filesystems for volume management.

# Create Container Flow

## Api post req to /namespaces/{namespace}/containers with container spec

## Api makes a new record in etcd for container

## Schedular assigns record to a node

## Node sees it has a container that is not running and ensures it matches the state

## First it checks if its networking matches the state, and create/removes rules

## Second it checks if its storage matches the state, and create/removes mounts

## Third it checks its containerd containers matches and creates/removes additional containers

## Finally it checks whether its containers match the status state and starts/stops any containers
