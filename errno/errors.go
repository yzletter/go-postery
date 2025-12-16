package errno

type Error struct {
	Code       int
	HTTPStatus int
	Msg        string
}

func (e *Error) Error() string { return e.Msg }

var (
	ErrServerInternal = &Error{50001, 500, "系统繁忙，请稍后重试"}
)
