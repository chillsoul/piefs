package storage

import (
	"github.com/pelletier/go-toml"
	"testing"
)

var config, _ = toml.LoadFile("../config.toml")
var (
	storeHost = config.Get("store.host")
	storePort = config.Get("store.port")
	storeDir  = config.Get("store.dir").(string)
)

func TestStore_GetNeedle(t *testing.T) {

}
