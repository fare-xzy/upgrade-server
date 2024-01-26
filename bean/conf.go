package bean

import (
	"github.com/BurntSushi/toml"
	"os"
)

type sysConfig struct {
	Servers []server `toml:"servers"`
}

type server struct {
	Name    string   `toml:"name"`
	Value   string   `toml:"value"`
	Paths   []string `toml:"paths"`
	Command string   `toml:"command"`
}

var (
	GlobalConfig sysConfig
)

func init() {
	_, _ = toml.DecodeFile("properties.toml", &GlobalConfig)

	// 获取可执行文件的路径
	_, err := os.Executable()
	if err != nil {
		panic(err)
	}
}
