package bean

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
	RunPath      string
)

func init() {
	//_, _ = toml.DecodeFile("properties.toml", &GlobalConfig)
	//
	//// 获取可执行文件的路径
	//_, err := os.Executable()
	//if err != nil {
	//	panic(err)
	//}
}
