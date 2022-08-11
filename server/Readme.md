# Server Implementation 

## Contents
- `Gossiper_server.go`: implementation of gossiper server functionality (i,e start server and handling gossip object requests)
- `Monitior_server.go`: implementation of monitor server functionality (i,e start server and handling requests)
- `Logger.go`: file that represents functionality of ctng logger
- `helper.go`: contains helper functions used within several other files

## Gossiper_server.go
- `bindContext`: binds context to functions that are passed to router
- `handleRequests`: starts http server to listen and serve requests
- `homePage`: base page of gossiper 
- The remaining functions are tailored towards handling different gossip requests such as POST and GET requests. The gossiper server start function is also contained in this file. 


## Monitor_server.go
- Contains Get functions for: `STH, Revocation(s), and POM`. 
- Receive functions for: `Gossip(object) and POM`.
- Handle functions for: `Gossip`
- After running `StartMonitorServer` the monitor will listen for requests by running `handleMonitorRequests`, then it runs the respective functions to handle the types of request listed above. 

## Logger.go
- This file represents the functionality of a CTng logger. Currently we use a fakelogger but this file is the skeleton for what a logger in CTng should look like. 


## helpers.go
- `gossipIDFromParams`: Gets and returns the Gossip_object_ID if it's valid, error if not. Gossip_object_ID is a struct which contains the values: application, type, signer, and period of a gossip object.
- `getSenderURL`: This gets and returns the senderURL, we use this in several functions to show where objects come from (valid, invalid, dups). 
- `isOwner`: Boolean function that returns true if the ownerURL matches the parsedURL, false if not. This is used in Gossip_server to verify if the sender is an owner.
