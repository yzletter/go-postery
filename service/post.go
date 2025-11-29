package service

import repository "github.com/yzletter/go-postery/repository/post"

type PostService struct {
	PostRepository repository.PostRepository
}

func NewPostService(postRepository repository.PostRepository) *PostService {
	return &PostService{PostRepository: postRepository}
}
