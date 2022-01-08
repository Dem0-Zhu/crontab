package master

import (
	"encoding/json"
	"io/ioutil"
)

// 配置文件解析类

var (
	G_config *Config
	i        int
)

type Config struct {
	ApiPort         int      `json:"api_port"`
	ApiReadTimeout  int      `json:"api_read_timeout"`
	ApiWriteTimeout int      `json:"api_write_timeout"`
	EtcdHosts       []string `json:"etcd_hosts"`
	EtcdDialTimeout int      `json:"etcd_dial_timeout"`
	WebRoot         string   `json:"webroot"`
}

func InitConfig(fileName string) (err error) {
	var (
		file []byte
		conf Config
	)
	if file, err = ioutil.ReadFile(fileName); err != nil {
		return
	}

	if err = json.Unmarshal(file, &conf); err != nil {
		return
	}

	// 配置单例
	G_config = &conf

	return
}
