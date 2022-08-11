# Gossiper Network Test

These folders specify a test network of four gossipers.


### Connections
Gossiper Connections are as follows:
* 1 - 2,3
* 2 - 3,4
* 3 - 4,1
* 4 - 1,2

Each Gossiper connects to the corresponding monitor number in 
monitorNetworkTest:
* 1-1
* 2-2
* 3-3
* 4-4

These servers are intended to provide a local test of all running components. The gossiper network can also be tested without monitors, although data must be sent manually + appropriately utilizing `client_test.go` to push data into gossiper 3.


### Running a gossiper
To run a gossiper, go to the root of this project and run `sh ./gossiperTest N` where N is the number of the gossiper you want to run. This runs ctng.go with the parameters from these test files.

## Code ownership:
Finn:
    - Made Configs and client test.

##### Readme Written by Finn