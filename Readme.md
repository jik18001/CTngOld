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

# Running the code

Run `go install .` before continuing!

To run:

**`a logger`**:  sh ./loggerTest.sh [loggerID]  

**`a monitor`**: sh ./monitorTest.sh [MonitorID]  

**`a gossiper`**: sh ./gossiperTest.sh [GossiperID]  

**`a CA`**:  sh ./CATest.sh [Certificate Authority ID]  

The `testData` folder contains configurations for testing, but configs can be generated using the functions in `config`.


# Function Documentation
Documentation + Function descriptions exist in each file/subfolder.

To view this this info + documentation in a formal documentation setting, GoDoc could be utilized, but requires installing the repository locally as a package.

### Licensing
Both imports we use, gorilla/mux and herumi/bls-go-binary, use an OpenBSD 3-clause license. as a result, we use the same Please see LICENSE in the outer folder for details.

##### Written By Finn and Jie
