package rag

import (
	"context"
	"fmt"
	"testing"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraMilvus "github.com/yzletter/go-postery/backend/infra/database/milvus"
	infraLLM "github.com/yzletter/go-postery/backend/infra/llm"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/tokenizer"
)

func TestRetrieve(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, "go_postery_test"+"/")
	// 加载私有配置
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"_test/")

	// 初始化千问 Embedder
	embedder := infraLLM.NewQwenEmbedder(ctx, InterviewServiceConf.Qwen)

	infraSlog.InitSlog(InterviewServiceConf.Log) // Init Slog

	// 初始化 Milvus Client
	MilvusClient := infraMilvus.NewMilvusClient(ctx, CommonMicroConf.Milvus)

	// Tokenizer
	Tokenizer := tokenizer.NewSegoTokenizer()

	// Milvus
	RAGDAO := NewMilvusRAGStore(MilvusClient, embedder, 10)

	// BM25
	bm25Retriever := NewBM25Retriever(Tokenizer, 10)

	// 加载题库到 Milvus
	err := RAGDAO.LoadQuestionsFromFile(ctx, 1, "test_data/test_question.json")
	if err != nil {
		fmt.Println("Milvus 加载题库失败")
		t.Failed()
		return
	}

	fmt.Println("Milvus 加载题库成功")

	// 加载题库到 BM25
	BM25Docs, err := LoadBM25DocsFromFile("test_data/test_question.json") // 从文件读取 转为 BM25Doc
	if err != nil {
		fmt.Println("BM25 转换文件失败")
		t.Failed()
		return
	}
	fmt.Println("BM25 转换文件成功")
	bm25Retriever.IndexBM25Docs(BM25Docs)

	// Milvus 检索
	docs, err := RAGDAO.RetrieveByUser(ctx, 1, "Go 并发编程")
	if err != nil {
		fmt.Printf("Milvus 召回失败\n%s\n", err)
		t.Failed()
		return
	}

	fmt.Println("Milvus 召回成功")
	for _, doc := range docs {
		fmt.Println(doc.ID, doc.Content, doc.MetaData)
	}

	// BM25 检索
	res, err := bm25Retriever.Retrieve(ctx, "Goroutine 线程 区别")
	if err != nil {
		fmt.Printf("BM25 召回失败\n%s\n", err)
		t.Failed()
	}
	fmt.Println("BM25 召回成功")
	for _, doc := range res {
		fmt.Println(doc.ID, doc.Content, doc.MetaData)
	}
}

// go test -v ./backend/micro/interview/rag -run=^TestRetrieve -count=1
