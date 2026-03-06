package index_service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"log/slog"

	model2 "github.com/yzletter/go-postery/microservice-backend/search/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Sentinel 分布式部署哨兵
type Sentinel struct {
	hub      ServiceHubInterface // 获取 IndexServiceWorker, 可以直接走 ServiceHub 也可以走 ServiceHubProxy
	connPool sync.Map            // grpc 连接池
}

// AddDoc 根据负载均衡算法选中一个 Worker 进行添加文档
func (sentinel *Sentinel) AddDoc(doc *model2.Document) (int, error) {
	// 根据负载均衡算法选中一个 Worker
	endpoint := sentinel.hub.GetServiceEndpoint(INDEX_SERVICE)
	if len(endpoint) <= 0 {
		return 0, errors.New("Get Endpoint Failed")
	}

	// 获取连接
	conn := sentinel.getConn(endpoint)
	if conn == nil {
		return 0, errors.New("Get grpc Connection Failed")
	}

	// 构造 grpc 客户端
	client := NewIndexServiceClient(conn)
	affected, err := client.AddDoc(context.Background(), doc)
	if err != nil {
		slog.Error("Worker Delete Document Failed", "endpoint", endpoint, "error", err)
		return 0, errors.New("Add Doc Failed")
	}

	return int(affected.Count), nil
}

// UpdateDoc 更新文档, 删除旧文档, 添加新文档
func (sentinel *Sentinel) UpdateDoc(doc *model2.Document) (int, error) {
	sentinel.DeleteDoc(doc.DocID)
	return sentinel.AddDoc(doc)
}

// DeleteDoc 删除文档, 向所有 Worker 发送删除请求
func (sentinel *Sentinel) DeleteDoc(docID string) int {
	// 获取所有 Endpoint
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) <= 0 {
		return 0
	}

	var cnt int32
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))

	// 开启协程异步从每个 Worker 上删除文档, 正常情况下只有一个 Worker 上有该文档
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			// 获取连接
			conn := sentinel.getConn(endpoint)
			if conn == nil {
				return
			}

			// 构造 grpc 客户端
			client := NewIndexServiceClient(conn)
			affected, err := client.DeleteDoc(context.Background(), &DocID{DocID: docID})
			if err != nil {
				slog.Error("Worker Delete Document Failed", "endpoint", endpoint, "error", err)
				return
			}
			if affected.Count > 0 {
				atomic.AddInt32(&cnt, affected.Count)
			}
		}(endpoint)
	}

	wg.Wait()
	return int(cnt)
}

// Search 向所有 Worker 发出搜索请求, 最后合并
func (sentinel *Sentinel) Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*model2.Document {
	// 获取所有 Endpoint
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) <= 0 {
		return nil
	}

	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))

	res := make([]*model2.Document, 0, 1000)
	resultCh := make(chan *model2.Document, 1000)

	// 开启协程异步从每个 Worker 上进行计数
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			// 获取连接
			conn := sentinel.getConn(endpoint)
			if conn == nil {
				return
			}

			// 构造 grpc 客户端
			client := NewIndexServiceClient(conn)
			searchResults, err := client.Search(context.Background(), &SearchRequest{
				TermQuery: query,
				OnFlag:    onFlag,
				OffFlag:   offFlag,
				OrFlags:   orFlags,
			})
			if err != nil {
				slog.Error("Worker Delete Document Failed", "endpoint", endpoint, "error", err)
				return
			}
			if len(searchResults.Results) > 0 {
				for _, result := range searchResults.Results {
					resultCh <- result
				}
			}
		}(endpoint)
	}

	allFinish := make(chan struct{}, 0)
	go func() {
		for {
			doc, ok := <-resultCh
			// 管道不可读
			if !ok {
				break
			}
			res = append(res, doc)
		}
		allFinish <- struct{}{} //3
	}()

	wg.Wait()
	close(resultCh) // 协程全部执行完后关闭管道入口
	<-allFinish     // 等待管道读完后返回结果
	return res
}

// Count 向所有 Worker 发出计数请求, 最后求和
func (sentinel *Sentinel) Count() int {
	// 获取所有 Endpoint
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) <= 0 {
		return 0
	}

	var cnt int32
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))

	// 开启协程异步从每个 Worker 上进行计数
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			// 获取连接
			conn := sentinel.getConn(endpoint)
			if conn == nil {
				return
			}

			// 构造 grpc 客户端
			client := NewIndexServiceClient(conn)
			affected, err := client.Count(context.Background(), &CountRequest{})
			if err != nil {
				slog.Error("Worker Delete Document Failed", "endpoint", endpoint, "error", err)
				return
			}
			if affected.Count > 0 {
				atomic.AddInt32(&cnt, affected.Count)
			}
		}(endpoint)
	}

	wg.Wait()
	return int(cnt)
}

// NewSentinel 构造函数
func NewSentinel(etcdServers []string) *Sentinel {
	return &Sentinel{
		//hub: GetServiceHub(etcdServers, 10), // 走 ServiceHub
		hub:      GetServiceHubProxy(etcdServers, 10, 100), // 走 ServiceHubProxy
		connPool: sync.Map{},
	}
}

// Close 关闭各个 grpc client connection 和 etcd client connection
func (sentinel *Sentinel) Close() (err error) {
	sentinel.connPool.Range(
		func(key, value any) bool {
			conn := value.(*grpc.ClientConn)
			err = conn.Close()
			return true
		})
	sentinel.hub.Close()
	return
}

// 从连接池中获取或复用连接
func (sentinel *Sentinel) getConn(endpoint string) *grpc.ClientConn {
	// 尝试复用连接
	if value, exist := sentinel.connPool.Load(endpoint); exist {
		conn := value.(*grpc.ClientConn)
		// 检查连接状态
		if conn.GetState() == connectivity.Connecting || conn.GetState() == connectivity.Ready {
			return conn
		}
		// 从池子中删除不可用连接
		conn.Close()
		sentinel.connPool.Delete(endpoint)
	}

	// 新建连接
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Credential即使为空，也必须设置
		// grpc.WithBlock(),
		// grpc.Dial是异步连接的，连接状态为正在连接。但如果你设置了 grpc.WithBlock 选项，就会阻塞等待（等待握手成功）。另外你需要注意，当未设置 grpc.WithBlock 时，ctx 超时控制对其无任何效果。
	)
	if err != nil {
		slog.Error("New grpc Connection Failed", "error", err)
		return nil
	}
	sentinel.connPool.Store(endpoint, conn)
	return conn
}
