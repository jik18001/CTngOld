# Next Generation Certificate Transparency:

## Folders:

**`config`**: Contains the layout of private and public json configuration files used by Monitors and Gossipers. Also contains loaders for the object versions of these .json files.

**`crypto`**: Abstractions associated with CTng cryptographic implementations, and the cryptoconfig implementation.

**`gossip`**: Defines the Gossip object, the state object of the Gossiper, and many functions associated with running a gossiper.

**`GZip`**: Implements functions for reading arrays of bytes to a base-64 encoded, Gzip-compressed representation.

**`monitor`**: Defines the monitor server state and functions associated with monitor tasks, such as querying Loggers and CAs.

**`server`**: HTTP Server implementations for the Gossiper and Monitor. Should call functions from **`gossip`** or **`monitor`**, respectively.

**`util`**: a package that has no internal imports: helper functions and constants that are used throughout the codebase but prevents import cycles from occurring (import cycles are not allowed in go).


**`testData`**: Defines a configuration of CTng with 4 monitors, 4 gossipers, 3 CAs, and 3 Loggers. Also defines a fakeLogger and fakeCA HTTP client for testing.
___
## Files

**`ctng.go`**: A commandline interface for running a Monitor, Gossiper, fakeLogger, or fakeCA.

**`monitorTest.sh`**: run as `sh monitorTest.sh [N]`, where N is which # monitor you want to run from the testData folder

**`gossiperTest.sh`**: run as `sh gossiperTest.sh [N]`, where N is which # monitor you want to run from the testData folder

# Running the code

Run `go install .` before continuing!

To run:

**`a logger`**  sh ./loggertest.sh [loggerID]  

**`a monitor`** sh ./monitortest.sh [MonitorID]  

**`a gossiper`** sh ./gossipertest.sh [GossiperID]  

**`a CA`**  sh ./CA.sh [Certificate Authority ID]  

The `testData` folder contains configurations for testing, but configs can be generated using the functions in `config`.


# Function Documentation
Documentation + Function descriptions exist in each file/subfolder.

To view this this info + documentation in a formal documentation setting, GoDoc could be utilized, but requires installing the repository locally as a package.

If you're having trouble viewing function descriptions and/or go language test in your IDE when opening this repository, try opening the CTng folder with your IDE instead of the outer SDP--PKI--CyberSec folder.

### Licensing
Both imports we use, gorilla/mux and herumi/bls-go-binary, use an OpenBSD 3-clause license. as a result, we use the same Please see LICENSE in the outer folder for details.

##### Written By Finn and Jie
