package directory

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"testing"
)

var config, _ = toml.LoadFile("../../config.toml")

func TestNewRedisDirectory(t *testing.T) {
	var redisDir, _ = NewRedisDirectory(config)
	pong, err := redisDir.db.Ping(ctx).Result()
	if err != nil {
		t.Error("error ping")
	}
	fmt.Println(pong)
}
func TestRedisDirectory_Get(t *testing.T) {
	var redisDir, _ = NewRedisDirectory(config)
	val, err := redisDir.db.Get(ctx, "test").Result()
	if err != nil {
		t.Error("error get", err)
	}
	fmt.Println(val)
}
