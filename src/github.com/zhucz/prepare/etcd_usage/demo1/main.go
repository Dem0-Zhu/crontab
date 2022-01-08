package main

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

/*
 创建一个etcd客户端
*/
func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
	)
	// 客户端配置
	config = clientv3.Config{
		Endpoints:   []string{"192.168.5.15:2379"},
		DialTimeout: 5 * time.Second,
	}
	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	client = client
}
