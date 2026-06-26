package index_service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"

	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
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
		return 0, errors.New("get search worker endpoint failed")
	}

	// 获取连接
	conn := sentinel.getConn(endpoint)
	if conn == nil {
		return 0, errors.New("get search worker grpc connection failed")
	}

	// 构造 grpc 客户端
	client := NewIndexServiceClient(conn)
	affected, err := client.AddDoc(context.Background(), doc)
	if err != nil {
		slog.Error("add document on search worker failed", "endpoint", endpoint, "doc_id", doc.DocID, "error", err)
		return 0, errors.New("add search document failed")
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
				slog.Error("delete document on search worker failed", "endpoint", endpoint, "doc_id", docID, "error", err)
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

	// 开启协程异步从每个 Worker 上进行搜索
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
				slog.Error("search on index worker failed", "endpoint", endpoint, "error", err)
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
		allFinish <- struct{}{}
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
				slog.Error("count documents on search worker failed", "endpoint", endpoint, "error", err)
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

// Close 关闭所有 gRPC Client 连接和 etcd 连接
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
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Credential 即使为空也必须设置
		my_grpc.CircuitBreakerDialOption(),
		grpc.WithKeepaliveParams(ka),
		// grpc.WithBlock(),
		// grpc.Dial 是异步连接的, 未设置 grpc.WithBlock 时 ctx 超时控制不会生效
	)
	if err != nil {
		slog.Error("create search worker grpc connection failed", "endpoint", endpoint, "error", err)
		return nil
	}
	sentinel.connPool.Store(endpoint, conn)
	return conn
}
