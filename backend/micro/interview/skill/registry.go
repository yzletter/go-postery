/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI Agent知识进阶，后端、AI大模型、场景题面试大全：https://golangstar.cn/
 */

package skill

// SkillRegistry Skill 注册中心，管理所有已注册的 Skill 并提供意图匹配
type SkillRegistry struct {
	skills []Skill // 已注册的 Skill 列表（按优先级排序，先注册优先级高）
}

// NewSkillRegistry 创建空的注册中心
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{}
}

// Register 注册一个 Skill（先注册的优先级高）
func (r *SkillRegistry) Register(s Skill) {
	r.skills = append(r.skills, s)
}

// Match 根据用户输入匹配合适的 Skill，遍历所有已注册的 Skill，返回第一个匹配的
func (r *SkillRegistry) Match(input string) Skill {
	for _, s := range r.skills {
		if s.Match(input) {
			return s
		}
	}
	return nil
}

// List 返回所有已注册的 Skill（用于帮助信息展示）
func (r *SkillRegistry) List() []Skill {
	return r.skills
}
