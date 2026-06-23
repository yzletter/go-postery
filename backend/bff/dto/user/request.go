package user

type ModifyProfileRequest struct {
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`   // 头像 URL
	Bio      string `json:"bio,omitempty"`      // 个性签名
	Gender   int    `json:"gender,omitempty"`   // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay string `json:"birthday,omitempty"` // 生日
	Location string `json:"location,omitempty"` // 地区
	Country  string `json:"country,omitempty"`  // 国家
}

type UploadCallbackRequest struct {
	Bucket string `json:"bucket"`
	Size   string `json:"size"`
	Object string `json:"object"`
}

type GetAvatarURLRequest struct {
	Avatar string `json:"avatar"`
}
