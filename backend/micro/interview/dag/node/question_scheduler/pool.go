package question_scheduler

import "github.com/yzletter/go-postery/backend/micro/interview/domain"

type questionPool struct {
	buckets map[domain.DifficultyLevel][]domain.PlannedQuestion
	remain  int
	backoff map[domain.DifficultyLevel][]domain.DifficultyLevel // 不同难度对应的退避顺序
}

func newQuestionPool(questions []domain.PlannedQuestion) *questionPool {
	p := &questionPool{
		buckets: map[domain.DifficultyLevel][]domain.PlannedQuestion{
			domain.DifficultyEasy:   make([]domain.PlannedQuestion, 0),
			domain.DifficultyMedium: make([]domain.PlannedQuestion, 0),
			domain.DifficultyHard:   make([]domain.PlannedQuestion, 0),
		},
		remain: 0,
		backoff: map[domain.DifficultyLevel][]domain.DifficultyLevel{
			domain.DifficultyEasy:   {domain.DifficultyEasy, domain.DifficultyMedium, domain.DifficultyHard},
			domain.DifficultyMedium: {domain.DifficultyMedium, domain.DifficultyEasy, domain.DifficultyHard},
			domain.DifficultyHard:   {domain.DifficultyHard, domain.DifficultyMedium, domain.DifficultyEasy}}}

	for _, q := range questions {
		level := q.Difficulty
		if _, ok := p.buckets[level]; !ok {
			level = domain.DifficultyMedium
		}
		p.buckets[level] = append(p.buckets[level], q)
		p.remain++
	}

	return p
}

func (pool *questionPool) next(target domain.DifficultyLevel) (question domain.PlannedQuestion, ok bool) {
	// 兜底
	backoff, ok := pool.backoff[target]
	if !ok {
		backoff = pool.backoff[domain.DifficultyMedium] // 传入 target 非法
	}

	// 遍历
	for _, level := range backoff {
		bucket := pool.buckets[level]
		if bucket == nil || len(bucket) == 0 { // 当前桶里没有
			continue
		}

		question := bucket[0]
		pool.buckets[level] = pool.buckets[level][1:] // 从桶里删除
		pool.remain--
		return question, true
	}

	return domain.PlannedQuestion{}, false
}

// empty 判断 pool 是否为空
func (pool *questionPool) empty() bool {
	return pool.remain == 0
}
