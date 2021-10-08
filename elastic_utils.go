package main

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"time"
)

var host = []string{
	//"http://118.195.194.75:32092/",
	//"http://1.117.11.230:9200/",
	"http://118.195.193.92:9200/",
}

// GetEsClient 初始化客户端
func GetEsClient(data Data) *elastic.Client {
	Init()
	file := "./eslog.log"
	logFile, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766) // 应该判断error，此处简略
	client, err := elastic.NewClient(
		elastic.SetURL(data.Host),
		elastic.SetSniff(false),
		elastic.SetInfoLog(log.New(logFile, "ES-INFO: ", 0)),
		elastic.SetTraceLog(log.New(logFile, "ES-TRACE: ", 0)),
		elastic.SetErrorLog(log.New(logFile, "ES-ERROR: ", 0)),
	)
	if err != nil {
		return nil
	}
	return client
}

// IsDocExists 判断索引是否存储
func IsDocExists(data Data, id string, index string) bool {
	client := GetEsClient(data)
	defer client.Stop()
	exist, _ := client.Exists().Index(index).Id(id).Do(context.Background())
	if !exist {
		log.Println("ID may be incorrect! ", id)
		return false
	}
	return true
}

// PingNode 是否联通
func PingNode(data Data) {
	start := time.Now()
	client := GetEsClient(data)
	info, code, err := client.Ping(data.Host).Do(context.Background())
	if err != nil {
		fmt.Printf("ping es failed, err: %v", err)
	}
	duration := time.Since(start)
	fmt.Printf("cost time: %v\n", duration)
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
}

// GetDoc 获取文档
func GetDoc(data Data, id string, index string) (*elastic.GetResult, error) {
	client := GetEsClient(data)
	defer client.Stop()
	if !IsDocExists(data, id, index) {
		return nil, fmt.Errorf("id不存在")
	}
	esResponse, err := client.Get().Index(index).Id(id).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return esResponse, nil
}

// CreateIndex 创建 index
func CreateIndex(data Data, index, mapping string) bool {
	client := GetEsClient(data)
	result, err := client.CreateIndex(index).BodyString(mapping).Do(context.Background())
	if err != nil {
		fmt.Printf("create index failed, err: %v\n", err)
		return false
	}
	return result.Acknowledged
}

// DelIndex 删除 index
func DelIndex(data Data, index ...string) bool {
	client := GetEsClient(data)
	response, err := client.DeleteIndex(index...).Do(context.Background())
	if err != nil {
		fmt.Printf("delete index failed, err: %v\n", err)
		return false
	}
	return response.Acknowledged
}

func getIndex(data Data) map[string]interface{} {
	client := GetEsClient(data)
	mapping := client.GetMapping()
	service := mapping.Index("*")
	result, err := service.Do(context.Background())
	if err != nil {
		fmt.Printf("create index failed, err: %v\n", err)
		return nil
	}
	return result
}
