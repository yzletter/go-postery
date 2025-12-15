package repository

import "github.com/yzletter/go-postery/model"

type postRepository struct {
}

func (repo *postRepository) Create(uid int, title, content string) (model.Post, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) Delete(pid int) error {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) Update(pid int, title, content string) error {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) GetByID(pid int) (bool, model.Post) {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) GetByPage(pageNo, pageSize int) (int, []model.Post) {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) GetByPageAndTag(tid, pageNo, pageSize int) (int, []model.Post) {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) GetByUid(uid int) []model.Post {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) ChangeViewCnt(pid int, delta int) {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) ChangeLikeCnt(pid int, delta int) {
	//TODO implement me
	panic("implement me")
}

func (repo *postRepository) ChangeCommentCnt(pid int, delta int) {
	//TODO implement me
	panic("implement me")
}
