package mq

import (
	infraRabbit "github.com/yzletter/go-postery/infra/rabbitmq"
	"github.com/yzletter/go-postery/service/ports"
)

// 用 RabbitMQ 实现 SessionMQ
type rabbitSessionMQ struct {
	mq *infraRabbit.RabbitMQ
}

func NewRabbitSessionMQ(mq *infraRabbit.RabbitMQ) ports.SessionMQ {
	return &rabbitSessionMQ{
		mq: mq,
	}
}
