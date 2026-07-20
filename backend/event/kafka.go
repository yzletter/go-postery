package event

import "github.com/bytedance/sonic"

const (
	// =================== Search Service ===================

	// KafkaSearchTopic 新文章建立索引 Kafka 消息 Topic
	KafkaSearchTopic = "search_index_document"
	// KafkaSearchGroup 新文章建立索引 Kafka 消息 Group
	KafkaSearchGroup = "search"

	// =================== Session Service ===================

	// KafkaSessionTopic 用户 IM 注册 Kafka 消息 Topic
	KafkaSessionTopic = "session_register_user"
	// KafkaSessionGroup 用户 IM 注册 Kafka 消息 Group
	KafkaSessionGroup = "session"

	// =================== Interactive Service ===================

	// KafkaTopicInteractiveRead 用户新阅读 Kafka 消息 Topic
	KafkaTopicInteractiveRead = "interactive_new_read"
	// KafkaTopicInteractiveLike 用户新点赞 Kafka 消息 Topic
	KafkaTopicInteractiveLike = "interactive_new_like"
	// KafkaTopicInteractiveFollow 用户新关注 Kafka 消息 Topic
	KafkaTopicInteractiveFollow = "interactive_new_follow"
	// KafkaTopicInteractiveComment 用户新评论 Kafka 消息 Topic
	KafkaTopicInteractiveComment = "interactive_new_comment"
	// KafkaInteractiveGroup 用户新互动 Kafka 消息消费 Group
	KafkaInteractiveGroup = "interactive"

	// =================== Rank Service ===================

	// KafkaTopicRankUpdateScore 更新分数 Kafka 消息 Topic
	KafkaTopicRankUpdateScore = "rank_update_score"
	// KafkaRankGroup 更新分数 Kafka 消息 Group
	KafkaRankGroup = "rank"
)

func NewKafkaOutboxEvent(id int64, topic string, key string, payload any) *OutboxEvent {
	value := ""
	switch v := payload.(type) {
	case string:
		value = v
	case []byte:
		value = string(v)
	default:
		value, _ = sonic.MarshalString(v)
	}

	return &OutboxEvent{
		ID:           id,
		Topic:        topic,
		MessageKey:   key,
		MessageValue: value,
	}
}
