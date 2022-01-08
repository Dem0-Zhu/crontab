package worker

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
	EtcdHosts       []string `json:"etcd_hosts"`
	EtcdDialTimeout int      `json:"etcd_dial_timeout"`
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
