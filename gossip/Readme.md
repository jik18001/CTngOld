# Gossip Package

## Contents
- `process_object.go`: Functions for processing a new object (valid, invalid, or duplicate)
- `gossiper.go`: Functions for actions that the Gossiper can complete as a client
  - Sending to Owner
  - Gossiping to connections
  - Accusing (unused currently)
- `gossip_object.go`: Functions for working with Gossip Objects
-   `accusations.go`: Describes the system for keeping track of accusations of each entity (with `accusation_validation.go` calls).
- Note that many calls are made between these files and the HTTP server. in the future, more gossiper logic could be moved from the server package to this one.

## `Types.go`
- Defines the Gossip Object and explains some design choices with Gossip_object_IDs.
  - Defines constants for the field types of CTng
- Defines the Gossiper Context object for managing server state

## Code Ownership:
Generally:
- Jie - Accusations handling + PoM Generation
- Isaac - Gossip Object Validation
- Finn - Gossip object type and processing functions
Individual function ownership is described in each file.

##### Readme Written by Finn