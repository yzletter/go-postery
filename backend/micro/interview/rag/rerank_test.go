package rag

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraMilvus "github.com/yzletter/go-postery/backend/infra/database/milvus"
	infraLLM "github.com/yzletter/go-postery/backend/infra/llm"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/tokenizer"
)

const TopK = 5

func TestRerank(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, "go_postery_test"+"/")
	// 加载私有配置
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"_test/")

	// 初始化千问大模型
	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen)

	// 初始化千问 Embedder
	embedder := infraLLM.NewQwenEmbedder(ctx, InterviewServiceConf.Qwen)

	infraSlog.InitSlog(InterviewServiceConf.Log) // Init Slog

	// 初始化 Milvus Client
	MilvusClient := infraMilvus.NewMilvusClient(ctx, CommonMicroConf.Milvus)

	// Tokenizer
	Tokenizer := tokenizer.NewSegoTokenizer()

	// Milvus
	RAGDAO := NewMilvusRAGStore(MilvusClient, embedder, TopK)

	// BM25
	BM25Retriever := NewBM25Retriever(Tokenizer, TopK)

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
	BM25Retriever.IndexBM25Docs(BM25Docs)

	retrievers := make([]retriever.Retriever, 0)
	retrievers = append(retrievers, BM25Retriever)
	retrievers = append(retrievers, RAGDAO)

	// 多路召回
	query := "Go 语言的 Goroutine 调度机制"
	MultiRetriever := NewMultiRetriever(retrievers, TopK)
	fusedRes, err := MultiRetriever.Retrieve(ctx, query)
	if err != nil {
		fmt.Println("多路召回失败")
		t.Failed()
		return
	}
	fmt.Println("多路召回成功")
	printRes(fusedRes)

	// 重排
	crossEncoderReranker := NewCrossEncoderReranker(InterviewServiceConf.Qwen.APIKey, "qwen3-vl-rerank", TopK)
	res, err := crossEncoderReranker.Rerank(ctx, query, fusedRes)
	if err != nil {
		fmt.Println("CrossEncoderReranker 重排失败")
		t.Failed()
		return
	}
	fmt.Println("CrossEncoderReranker 重排成功")
	printRes(res)

	scoreBasedReranker := NewScoreBasedReranker(TopK)
	res, err = scoreBasedReranker.Rerank(ctx, query, fusedRes)
	if err != nil {
		fmt.Println("ScoreBasedReranker 重排失败")
		t.Failed()
		return
	}
	fmt.Println("ScoreBasedReranker 重排成功")
	printRes(res)

	llmReranker := NewLLMReranker(QwenLLMModel, TopK)
	res, err = llmReranker.Rerank(ctx, query, fusedRes)
	if err != nil {
		fmt.Println("LLMReranker 重排失败")
		t.Failed()
		return
	}
	fmt.Println("LLMReranker 重排成功")
	printRes(res)

	noneReranker := NewNoneReranker(TopK)
	res, err = noneReranker.Rerank(ctx, query, fusedRes)
	if err != nil {
		fmt.Println("NoneReranker 重排失败")
		t.Failed()
		return
	}
	fmt.Println("NoneReranker 重排成功")
	printRes(res)
}

func printRes(res []*schema.Document) {
	for _, doc := range res {
		score, _ := doc.MetaData["_rrf_score"].(float64)
		fmt.Printf("文档\n%s\n分数%.4f\n", doc, score)
	}
	fmt.Println("===================================================================")
}

// go test -v ./backend/micro/interview/rag -run=^TestRerank -count=1
