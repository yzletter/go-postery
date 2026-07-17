package utils

import (
	"encoding/json"
	"testing"
)

func TestBindErrMsgReturnsNonValidationError(t *testing.T) {
	err := &json.SyntaxError{Offset: 1}

	if got := BindErrMsg(err); got != err.Error() {
		t.Fatalf("BindErrMsg() = %q, want %q", got, err.Error())
	}
}
