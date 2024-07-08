package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/crt379/svc-collector-grpc-gw/internal/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	EtcdClient *clientv3.Client
)

func init() {
	var err error

	endpoints := make([]string, len(config.AppConfig.Etcd))
	for i, a := range config.AppConfig.Etcd {
		endpoints[i] = fmt.Sprintf("%s:%s", a.Host, a.Port)
	}

	EtcdClient, err = NewEtcdClient(endpoints, 5*time.Second)
	if err != nil {
		log.Panicf(err.Error())
	}
}

func NewEtcdClient(endpoints []string, dialtimeout time.Duration) (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialtimeout,
	})
	if err != nil {
		return client, err
	}

	return client, err
}
