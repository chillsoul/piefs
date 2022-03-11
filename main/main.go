package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"piefs/master"
	"piefs/storage"
)

func main() {
	config, err := toml.LoadFile("./config.toml")
	if err != nil {
		fmt.Println("error load config:", err)
		return
	}
	s, err := storage.NewStorage(config)
	if err != nil {
		fmt.Println("error new storage:", err)
		return
	}
	m, err := master.NewMaster(config)
	if err != nil {
		fmt.Println("error new master:", err)
		return
	}
	go s.Start()
	go m.Start()
	select {}
}
