package worker

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

// Register 注册节点到etcd cron/worker/{IP}
type Register struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease

	localIP string // 本机IP
}

var G_register *Register

// 获取本机IP
func getLocalIP() (ipv4 string, err error) {
	var addrs []net.Addr
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	//取第一个非localhost（lo）的网卡
	for _, addr := range addrs {
		// 这个网络地址是ip地址，并且不是换回地址（127.0.0.1）
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			// 跳过ipv6
			if ip.IP.To4() != nil {
				ipv4 = ip.IP.String()
				return
			}
		}
	}
	return "", common.ErrNoLocalIpFound
}

func InitRegister() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		localIP string
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdHosts,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	if localIP, err = getLocalIP(); err != nil {
		return
	}
	G_register = &Register{
		client: client,
		kv: kv,
		lease: lease,
		localIP: localIP,
	}
	go G_register.keepOnline()
	return
}

// 注册到/cron/worker/{IP} 并自动续租
func (register *Register) keepOnline()  {
	var (
		regKey string
		err error
		leaseGrantResp *clientv3.LeaseGrantResponse
		keepAliveChan <- chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
		cancelTxn context.Context
		cancelFunc context.CancelFunc
	)
	// 注册路径
	regKey = common.JobWorkerDir + register.localIP
	for {
		cancelFunc = nil
		// 创建租约， 10s过期，删除key
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		// 自动续租
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			time.Sleep(1 * time.Second)
			goto RETRY
		}

		cancelTxn, cancelFunc = context.WithCancel(context.TODO())
		//注册到etcd
		if _, err = register.kv.Put(cancelTxn, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		for {
			select {
			case keepAliveResp = <- keepAliveChan:
				if keepAliveResp == nil {
					// 续租失败
					goto RETRY
				}
			}
		}

		RETRY:
			time.Sleep(1 * time.Second)
			if cancelFunc != nil {
				cancelFunc()
			}
	}
}



