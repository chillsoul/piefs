package main

import (
	"fmt"
	"github.com/chillsoul/piefs/master"
	"github.com/chillsoul/piefs/storage"
	"github.com/chillsoul/piefs/util/config"
)

func main() {
	config1, err := config.LoadConfig("./config1.toml")
	s1, err := storage.NewStorage(config1.Master, config1.Storage, config1.Cache)
	if err != nil {
		fmt.Println("error new storage:", err)
		return
	}
	//s2, err := storage.NewStorage(config2)
	if err != nil {
		fmt.Println("error new storage:", err)
		return
	}
	m, err := master.NewMaster(config1.Master)
	if err != nil {
		fmt.Println("error new master:", err)
		return
	}
	go m.Start()
	go s1.Start()
	//go s2.Start()
	select {}
}
