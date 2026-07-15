package service

import (
	"context"
)

type InterviewService interface {
	// Chat 发起面试辅助对话
	Chat(ctx context.Context, userID int64, input string) (string, error)

	// StartInterview 开始面试
	StartInterview(ctx context.Context, userID int64, jd string, resume string, candidateName string) (int64, error)

	// Answer 回答当前面试题
	Answer(ctx context.Context, userID int64, sessionID int64, answer string) error

	// Evaluation 生成面试评估报告
	Evaluation(ctx context.Context, userID int64, sessionID int64) error

	// UploadQuestionsSign 获取题库上传 OSS 签名
	UploadQuestionsSign(ctx context.Context, userID int64) (string, error)

	// UploadQuestionsCallback 处理题库上传回调
	UploadQuestionsCallback(ctx context.Context, userID int64, object string) error

	// UploadQuestions 上传用户题库
	UploadQuestions(ctx context.Context, userID int64, sourceFile string, data []byte) error

	// QuitInterview 退出面试
	QuitInterview(ctx context.Context, userID int64, sessionID int64) error
}
