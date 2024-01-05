# Plaid

Tooling to simplify local development in a multi-services and data exploration environments.  Plaid builds on top of
existing tools to concentrate on orchestration of services and systems.

## Examples

* [simple](tests/system/simple) is a project to run a single command.  This is the simplest example.
* [services](tests/system/deps/services) is an example of running a single command which is dependent upon another task.

# Releases
* 0.2.0 - Stabilizing race conditions and exploring state management.
* 0.1.0 - Initial proof of concept

# Developing
* Run `./test-daemon.sh`.  This will test the system then build binaries into `test/system` as `plaid-client` and
`plaid-daemon`.
* Run `./test/system/plaid-daemon run` in a terminal
* Run `./test/system/plaid-client up` in a directory to pull start.

Check out the running test cases under `test-daemon.sh` for more examples of how to run things!

# Roadmap
This is really just a list of features which I would like to implement at some point.

## Features
* Namespaces - Provides a method to avoid conflicting resources.  Similar to Kubernetes approach to isolation and security.
* Spec and Status validation - Having this enabled by default would be helpful in preventing some classes of dumb errors.
* Kind + version listing - Able to list which kinds exist within the system
* Kind output specs - Able to specify which fields are important in general output 
* Metadata - Modified time, created time, etc
* Annotations and labels - These are often helpful in Kubernetes.  I can foresee some usages would be great.
* Multi-phase delete - When deleting the resources should translate into deleting then deleted.
* Service Discovery && Connection - Able to describe to services how to connect to peers on the network and perhaps in the data storage arena.
  * Activation - Services are booted and activated upon need.  As a second phase activation based on versioned requirements.

## Optimizations
* `resources.ClientWatcher` does not propagate filters into the `resource.Controller` reactor, meaning all events are
dispatched to all `ClientWatcher`s which exist within the system.  Obviously does not scale but is functional for now.

## Refactoring
* `resources/client.go` contains a lot of concerns.  These should probably be broken out into the correct operations
file.
* `resources.Client` contains a lot of dispatching code which effectively boils down to adapting asynchronous calling
conventions.  This should be cleaned up to either user Junk Bucket's reactors or rethought to simplify.
