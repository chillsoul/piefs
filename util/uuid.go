package util

import "time"

func UniqueId() (id uint64) {
	//TODO 分布式唯一自增id snowflake
	//Will return same uuid when high concurrent calls
	return uint64(time.Now().UnixNano())
}
