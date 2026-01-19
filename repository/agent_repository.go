package repository

import "github.com/yzletter/go-postery/repository/dao"

type agentRepository struct {
	dao dao.AgentDAO
}

func NewAgentRepository(dao dao.AgentDAO) AgentRepository {
	return &agentRepository{dao: dao}
}
