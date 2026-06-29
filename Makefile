.PHONY: build build-macos

build:
	@rm -rf ./app/linux || true
	@go mod tidy
 	# 9001
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/code_service ./backend/micro/code
	# 9002
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/auth_service ./backend/micro/auth
	# 9003
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/lottery_service ./backend/micro/lottery
	# 9004
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/post_service ./backend/micro/post
	# 9005
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/search_service ./backend/micro/search
	# 9007
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/user_service ./backend/micro/user
	# 9008
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/session_service ./backend/micro/session
	# 9009
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/rank_service ./backend/micro/rank
	# 9010
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/interactive_service ./backend/micro/interactive
	# 无端口
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/outbox_service ./backend/event/outbox

	@scp ./app/linux/code_service production1:~/app/code_service
	@scp ./app/linux/code_service production2:~/app/code_service
	@scp ./app/linux/code_service production3:~/app/code_service
	# auth_service 9002
	@scp ./app/linux/auth_service production1:~/app/auth_service
	@scp ./app/linux/auth_service production2:~/app/auth_service
	@scp ./app/linux/auth_service production3:~/app/auth_service
	# lottery_service 9003
	@scp ./app/linux/lottery_service production1:~/app/lottery_service
	@scp ./app/linux/lottery_service production3:~/app/lottery_service
	# post_service 9004
	@scp ./app/linux/post_service production1:~/app/post_service
	@scp ./app/linux/post_service production2:~/app/post_service
	@scp ./app/linux/post_service production3:~/app/post_service
	# search_service 9005
	@scp ./app/linux/search_service production3:~/app/search_service
	# user_service 9007
	@scp ./app/linux/user_service production1:~/app/user_service
	@scp ./app/linux/user_service production2:~/app/user_service
	@scp ./app/linux/user_service production3:~/app/user_service
	# session_service 9008
	@scp ./app/linux/session_service production1:~/app/session_service
	@scp ./app/linux/session_service production2:~/app/session_service
	@scp ./app/linux/session_service production3:~/app/session_service
	# rank_service 9009
	@scp ./app/linux/rank_service production1:~/app/rank_service
	@scp ./app/linux/rank_service production2:~/app/rank_service
	# interactive_service 9010
	@scp ./app/linux/interactive_service production1:~/app/interactive_service
	@scp ./app/linux/interactive_service production2:~/app/interactive_service
	@scp ./app/linux/interactive_service production3:~/app/interactive_service
	# outbox 无端口
	@scp ./app/linux/outbox_service production3:~/app/outbox_service
build-macos:
	@mkdir -p ./app/macos
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/code_service ./backend/micro/code
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/auth_service ./backend/micro/auth
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/lottery_service ./backend/micro/lottery
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/post_service ./backend/micro/post
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/agent_service ./backend/micro/agent
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/user_service ./backend/micro/user
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/session_service ./backend/micro/session
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/outbox_service ./backend/micro/outbox
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/bff_service ./backend/micro/bff
