package event

const (
	// KafkaSearchTopic 新文章建立索引 Kafka 消息 Topic
	KafkaSearchTopic = "search_index_document"
	// KafkaSearchGroup 新文章建立索引 Kafka 消息 Group
	KafkaSearchGroup = "search_index_document"

	// KafkaSessionTopic 用户 IM 注册 Kafka 消息 Topic
	KafkaSessionTopic = "session_register_user"
	// KafkaSessionGroup 用户 IM 注册 Kafka 消息 Group
	KafkaSessionGroup = "session_register_user"

	// KafkaRAGTopic 新文章 RAG 切分 Chunk Kafka 消息 Topic
	KafkaRAGTopic = "rag_chunk_document"
	// KafkaRAGGroup 新文章 RAG 切分 Chunk Kafka 消息 Group
	KafkaRAGGroup = "rag_chunk_document"

	// KafkaQdrantTopic 新向量入 Qdrant Kafka 消息 Topic
	KafkaQdrantTopic = "agent_insert_vector"
	// KafkaQdrantGroup 新向量入 Qdrant Kafka 消息 Group
	KafkaQdrantGroup = "agent_insert_vector"

	// KafkaUserTopic 用户初始化排行榜 Kafka 消息 Topic
	KafkaUserTopic = "user_init_score"
	// KafkaUserGroup 用户初始化排行榜 Kafka 消息 Group
	KafkaUserGroup = "user_init_score"
)
