package index_service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"log/slog"

	"github.com/yzletter/go-postery/search/model"
	"github.com/yzletter/go-postery/search/utils"
)

const (
	INDEX_SERVICE = "index_service"
)

// IndexServiceWorker 实现 protoc 生成的 grpc Server
type IndexServiceWorker struct {
	Indexer    *Indexer
	serviceHub *ServiceHub // 所属注册中心
	selfAddr   string      // IP + Port
	UnimplementedIndexServiceServer
}

// Init 初始化索引
func (service *IndexServiceWorker) Init(DocNumEstimate int, DataDir string) error {
	service.Indexer = new(Indexer)
	return service.Indexer.Init(DocNumEstimate, DataDir)
}

// Register 向注册中心注册自己
func (service *IndexServiceWorker) Register(etcdServers []string, servicePort int) error {
	// 参数校验
	if len(etcdServers) <= 0 {
		return errors.New("Invalid Etcd Servers")
	} else if servicePort <= 1024 {
		return errors.New("Invalid Service Port")
	}

	// 获取本机内网 IP
	localIP, err := utils.GetLocalIP()
	if err != nil {
		slog.Error("Get Local IP Failed", "error", err)
		panic(err)
	}
	localIP = "127.0.0.1" // TODO 单机模拟分布式时，把 localIP 写死为 127.0.0.1

	service.selfAddr = localIP + ":" + strconv.Itoa(servicePort)
	var heartBeat int64 = 3                      // 每隔3秒上报一次心跳
	hub := GetServiceHub(etcdServers, heartBeat) // 单例
	leaseID, err := hub.Register(INDEX_SERVICE, service.selfAddr, 0)
	if err != nil {
		slog.Error("Index Service Register Failed", "error", err)
		panic(err)
	}
	service.serviceHub = hub

	// 启动协程进行续约
	go func() {
		for {
			hub.Register(INDEX_SERVICE, service.selfAddr, leaseID)
			time.Sleep(time.Duration(heartBeat)*time.Second - 100*time.Millisecond)
		}
	}()

	return nil
}

// Close 关闭索引
func (service *IndexServiceWorker) Close() error {
	if service.serviceHub != nil {
		service.serviceHub.Unregister(INDEX_SERVICE, service.selfAddr)
	}
	return service.Indexer.Close()
}

func (service *IndexServiceWorker) AddDoc(ctx context.Context, document *model.Document) (*AffectedCount, error) {
	n, err := service.Indexer.AddDoc(document)
	return &AffectedCount{Count: int32(n)}, err
}

func (service *IndexServiceWorker) DeleteDoc(ctx context.Context, id *DocID) (*AffectedCount, error) {
	n := service.Indexer.DeleteDoc(id.DocID)
	return &AffectedCount{Count: int32(n)}, nil
}

func (service *IndexServiceWorker) Search(ctx context.Context, request *SearchRequest) (*SearchResult, error) {
	res := service.Indexer.Search(request.TermQuery, request.OnFlag, request.OffFlag, request.OrFlags)
	return &SearchResult{Results: res}, nil
}

func (service *IndexServiceWorker) Count(ctx context.Context, request *CountRequest) (*AffectedCount, error) {
	n := service.Indexer.Count()
	return &AffectedCount{Count: int32(n)}, nil
}
