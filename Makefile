.PHONY: build build-macos

build:
	@rm -rf ./app/linux || true
	@go mod tidy
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/code_service ./backend/micro/code
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/lottery_service ./backend/micro/lottery
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/bff_service ./backend/micro/bff

	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/auth_service ./backend/micro/auth
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/post_service ./backend/micro/post
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/agent_service ./microservice-backend/agent
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/user_service ./microservice-backend/user
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/session_service ./microservice-backend/session
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/outbox_service ./backend/micro/outbox

	@scp ./app/linux/code_service production1:~/app/code_service
	@scp ./app/linux/code_service production2:~/app/code_service
	@scp ./app/linux/code_service production3:~/app/code_service

	@scp ./app/linux/lottery_service production1:~/app/lottery_service
	@scp ./app/linux/lottery_service production3:~/app/lottery_service

	@scp ./app/linux/outbox_service production3:~/app/outbox_service

	@scp ./app/linux/auth_service production1:~/app/auth_service
	@scp ./app/linux/agent_service production2:~/app/agent_service
	@scp ./app/linux/user_service production1:~/app/user_service
	@scp ./app/linux/session_service production1:~/app/session_services

build-macos:
	@mkdir -p ./app/macos
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/code_service ./microservice-backend/code
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/auth_service ./backend/micro/auth
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/lottery_service ./microservice-backend/lottery
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/post_service ./backend/micro/post
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/agent_service ./microservice-backend/agent
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/user_service ./microservice-backend/user
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/session_service ./microservice-backend/session
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/outbox_service ./backend/micro/outbox
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/bff_service ./microservice-backend/bff
