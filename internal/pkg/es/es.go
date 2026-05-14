package es

import (
	"fmt"
	"sync"

	"github.com/olivere/elastic/v7"
)

var (
	client *elastic.Client
	once   sync.Once
)

// Init 初始化Elasticsearch客户端
func Init(addresses []string, username, password string, sniff bool) error {
	var err error
	once.Do(func() {
		options := []elastic.ClientOptionFunc{
			elastic.SetURL(addresses...),
			elastic.SetSniff(sniff),
			elastic.SetHealthcheck(true),
		}
		if username != "" && password != "" {
			options = append(options, elastic.SetBasicAuth(username, password))
		}

		c, e := elastic.NewClient(options...)
		if e != nil {
			err = fmt.Errorf("elasticsearch connect failed: %w", e)
			return
		}
		client = c
	})
	return err
}

// GetClient 获取ES客户端
func GetClient() *elastic.Client {
	return client
}
