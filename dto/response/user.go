package dto

// UserResponse 后端返回
type UserResponse struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func GetUserResponse(id int, name string) UserResponse {
	return UserResponse{
		Id:   id,
		Name: name,
	}
}
