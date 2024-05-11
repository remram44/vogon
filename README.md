# Vogon - a scriptable orchestration platform

Don't read too much into it -- this is an experiment.

## What is the goal?

Kubernetes is a great model for cluster orchestration. A single database holds objects, with both their desired state and current status. Multiple layers of components implement control loops, reading the state of some objects and updating the state of other objects, possibly taking actions, but never having to worry about their interaction.

However Kubernetes is quite rigid, as controllers need to be full-fleged microservices which authenticate to the API, with specially provisioned service accounts, and use cluster-global resource definitions. Workloads are assumed to be uniform, with pods everywhere being the same and nodes being configured similarly. Namespaces exist for multi-tenancy but they cannot be nested, have to be created by an admin, and share resource types. Using multiple clusters in a single organization is common.

This project aims to be a more customizable, programmable, and stretchable platform. Objects in one namespace can be mapped to different objects elsewhere using scripts that are easy to author and managed by the platform.

The idea is to focus on the API integration aspect, building powerful primitives for turning JSON resources into other JSON resources according to easy but expressive sets of rules. Then build connectors to have those JSON resources come to life as VMs, Pods, clusters, etc. Messing with low-level runtimes components (kubelet, CRI, CNI, CSI) should be out of scope.

## Plan

- Implement database connectors
    - [x] In-memory
    - [ ] Local files
    - [ ] Etcd
- [x] CRUD API
    - [ ] Watch
- [ ] Lua integration
- [ ] Control loops in Lua
- [ ] Synchronous mutation hooks in Lua
- [ ] Kubernetes connector
- [ ] Resource schemas (e.g. validation against JSON schemas, version conversion)
- [ ] Authn / authz
