package main

/*
Code Ownership:
Finn - Made main function
*/
import (
	"CTng/config"
	"CTng/gossip"
	"CTng/monitor"
	"CTng/server"
	"CTng/testData/fakeCA"
	fakeLogger "CTng/testData/fakeLogger"
	"fmt"
	"os"
)

// main is run when the user types "go run ."
// it allows a user to run a gossiper, monitor, fakeLogger, or fakeCA.
// Currently unimplemented: Different object_storage locations than ./gossiper_data.json and ./monitor_data.json
// This field could be defined within the configuration files to make this more modular.
func main() {
	helpText := "Usage:\n ./CTng [gossiper|monitor] <public_config_file_path> <private_config_file_path> <crypto_config_path>\n ./Ctng [logger|ca] <fakeentity_config_path>"
	if len(os.Args) < 3 {
		fmt.Println(helpText)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "gossiper":
		// make the config object.
		conf, err := config.LoadGossiperConfig(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			fmt.Println(helpText)
			panic(err)
		}
		// Space is allocated for all storage fields, and then make is run to initialize these spaces.
		storage := new(gossip.Gossip_Storage)
		*storage = make(gossip.Gossip_Storage)
		accusationdb := new(gossip.AccusationDB)
		*accusationdb = make(gossip.AccusationDB)

		ctx := gossip.GossiperContext{
			Config:      &conf,
			Storage:     storage,
			Accusations: accusationdb,
			StorageFile: "gossiper_data.json", // could be a parameter in the future.
			HasPom:      make(map[string]bool),
		}
		ctx.Config = &conf
		server.StartGossiperServer(&ctx)
		// break // break unneeded in  go.
	case "monitor":
		// make the config object.
		conf, err := config.LoadMonitorConfig(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			fmt.Println(helpText)
			panic(err)
		}
		// Space is allocated for all storage fields, and then make is run to initialize these spaces.
		storage := new(gossip.Gossip_Storage)
		*storage = make(gossip.Gossip_Storage)
		ctx := monitor.MonitorContext{
			Config:      &conf,
			Storage:     storage,
			StorageFile: "monitor_data.json",
			StorageID:   os.Args[5],
			HasPom:      make(map[string]bool),
			HasAccused:  make(map[string]bool),
		}
		ctx.Config = &conf
		server.StartMonitorServer(&ctx)
	case "logger":
		fakeLogger.RunFakeLogger(os.Args[2])
	case "ca":
		fakeCA.RunFakeCA(os.Args[2])
	default:
		fmt.Println(helpText)
	}
}
