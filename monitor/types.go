package monitor

import (
	"CTng/config"
	"CTng/gossip"
	"CTng/util"
	"encoding/json"
	"net/http"
)

type MonitorContext struct {
	Config      *config.Monitor_config
	Storage     *gossip.Gossip_Storage
	StorageFile string
	StorageFile_accusations string
	StorageFile_PoMs string
	// TODO: Utilize Storage directory: A folder for the files of each MMD.
	// Folder should be set to the current MMD "Period" String upon initialization.
	StorageDirectory string
	StorageID string

	// The below could be used to prevent a Monitor from sending duplicate Accusations,
	// Currently, if a monitor accuses two entities in the same Period, it will trigger a gossip PoM.
	// Therefore, a monitor can only accuse once per Period. I believe this is a temporary solution.
	HasPom     map[string]bool
	HasAccused map[string]bool
	Verbose    bool
	Client     *http.Client
}

func (c *MonitorContext) SaveAccusations() error{
	err:= util.WriteData(c.StorageDirectory+"/"+c.StorageFile_accusations, c.HasAccused)
	return err
}

func (c *MonitorContext) LoadAccusations() error{
	bytes, err := util.ReadByte(c.StorageFile_accusations)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &c.HasAccused)
	if err != nil {
		return err
	}
	return nil
}

func (c *MonitorContext) SavePoMs() error{
	err:= util.WriteData(c.StorageDirectory+"/"+c.StorageFile_PoMs, c.HasPom)
	return err
}

func (c *MonitorContext) LoadPoMs() error{
	bytes, err := util.ReadByte(c.StorageFile_PoMs)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &c.HasPom)
	if err != nil {
		return err
	}
	return nil
}

func (c *MonitorContext) SaveStorage() error {
	storageList := []gossip.Gossip_object{}
	for _, gossipObject := range *c.Storage {
		storageList = append(storageList, gossipObject)
	}
	err := util.WriteData(c.StorageDirectory+"/"+c.StorageFile, storageList)
	return err
}

func (c *MonitorContext) Saveall(){
	c.SaveStorage()
	c.SaveAccusations()
	c.SavePoMs()
}

func (c *MonitorContext) LoadStorage() error {
	storageList := []gossip.Gossip_object{}
	bytes, err := util.ReadByte(c.StorageFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &storageList)
	if err != nil {
		return err
	}
	for _, gossipObject := range storageList {
		(*c.Storage)[gossipObject.GetID(int64(c.Config.Public.Gossip_wait_time))] = gossipObject
	}
	return nil
}

func (c *MonitorContext) GetObject(id gossip.Gossip_object_ID) (gossip.Gossip_object, bool) {
	obj := (*c.Storage)[id]
	if obj == (gossip.Gossip_object{}) {
		return obj, false
	}
	return obj, true
}
func (c *MonitorContext) IsDuplicate(g gossip.Gossip_object) bool {
	//no public period time for monitor :/
	id := g.GetID(int64(c.Config.Public.Gossip_wait_time))
	obj := (*c.Storage)[id]
	if obj == (gossip.Gossip_object{}) {
		return false
	}
	return true
}

func (c *MonitorContext) StoreObject(o gossip.Gossip_object) {
	(*c.Storage)[o.GetID(int64(c.Config.Public.Gossip_wait_time))] = o
}
