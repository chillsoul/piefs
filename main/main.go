package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"piefs/master"
	"piefs/storage"
)

func main() {
	config, err := toml.LoadFile("./config.toml")
	config2, err := toml.LoadFile("./config2.toml")
	s1, err := storage.NewStorage(config)
	if err != nil {
		fmt.Println("error new storage:", err)
		return
	}
	s2, err := storage.NewStorage(config2)
	if err != nil {
		fmt.Println("error new storage:", err)
		return
	}
	m, err := master.NewMaster(config)
	if err != nil {
		fmt.Println("error new master:", err)
		return
	}
	go m.Start()
	go s1.Start()
	go s2.Start()
	select {}
}
