package dto

import "github.com/yzletter/go-postery/model"

// UserDTO 后端返回
type UserDTO struct {
	Id   int    `json:"id,omitempty,string"`
	Name string `json:"name,omitempty"`
}

// ToUserDTO model.User 转 UserDTO
func ToUserDTO(user model.User) UserDTO {
	return UserDTO{
		Id:   user.Id,
		Name: user.Name,
	}
}

// ToUserDTOs []model.User 转 []UserDTO
func ToUserDTOs(users []model.User) []UserDTO {
	res := make([]UserDTO, len(users))
	for _, user := range users {
		res = append(res, ToUserDTO(user))
	}
	return res
}
