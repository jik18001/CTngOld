# Monitor Implementation 

## Contents
- `types.go`: type declarations for monitor context and some basic monitor methods definitions
- `monitor.go`: implementation of clientside monitor functions that will be used by monitor_server in the server folder
- `monitor_process.go`: contains only process valid gossip object functions

## types.go
-`Monitor_context`: monitor context is an object that contains all the configuration and storage information about the monitor
- `methods`: internal methods defined in this file includes savestorage, loadstorage, getobject, isduplicate, and storeobject 
## monitor.go
- `Queryloggers`: send HTTP get request to loggers
- `QueryAuthorities`: send HTTP get request to CAs
- `Check_entity_pom`: check if there is a pom against the provided URL 
- `isLogger`: check if the entities is in the Loggers list from the public config file
- `IsAuthority`: check if the entities is in the CAs list from the public config file
- `Check_entity_pom`: check if there is a PoM aganist this entity 
- `AccuseEntity`: accuses the entity if its URL is provided   
- `Send_to_gossiper`: send the input gossip object to the gossiper  
- `PeriodicTasks` : query loggers once per MMD, accuse if the logger is inactive
## monitor_process.go
- `Process_valid_object`:  
  If the monitor receives a valid sth from a logger or valid revocation information from a CA (both need to be in the form of a gossip object),  
  this function will send the raw version of the gossip object to the gossiper, wait for gossip_wait_time, check updated local PoM database,  
  if there is no PoM against the signer of this sth/revocation information, threshold sign this sth/revocation information and send it to the gossiper  


## Code Owners
- types.go is written by Finn and Marcus
- monitor_process.go is written by Jie, reviewed by Finn
- `Queryloggers` and `QueryAuthorities` functions in monitor.go are written by Marcus, reviewed and edited by Finn
- `AccuseEntity` and `Send_to_Gossiper` functions in monitor.go  are written by Jie, reviewed by Finn
- `Check_entity_pom` `isLogger` and `IsAuthority` functions monitor.go are written by Jie, reviewed and edited by Finn
- `PeriodicTasks` function in monitor.go is written by Finn
##### Readme written by Jie
