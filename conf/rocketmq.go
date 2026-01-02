package conf

import "time"

const (
	RocketLotteryTopic         = "GO_POSTERY_CANCEL_ORDER"
	RocketLotteryConsumerGroup = "go_postery"
	RocketProxyEndpoint        = "127.0.0.1:8081"
	RocketAwaitDuration        = 5 * time.Second
	RocketLotteryPayDelay      = 600
)

// sh mqadmin updateTopic -n localhost:9876 -c DefaultCluster -t GO_POSTERY_CANCEL_ORDER -a +message.type=DELAY
// sh mqadmin deleteTopic -n localhost:9876 -c DefaultCluster -t GO_POSTERY_CANCEL_ORDER
// sh mqadmin updateSubGroup -n localhost:9876 -c DefaultCluster -g go_postery
