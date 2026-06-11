.PHONY: build build-macos

build:
	@rm -rf ./app/linux || true
	@go mod tidy
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/code_service ./micro-backend/code
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/auth_service ./microservice-backend/auth
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/lottery_service ./micro-backend/lottery
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/post_service ./microservice-backend/post
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/agent_service ./microservice-backend/agent
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/user_service ./microservice-backend/user
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/session_service ./microservice-backend/session
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/outbox_service ./backend/micro/outbox
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/bff_service ./microservice-backend/bff

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
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/auth_service ./microservice-backend/auth
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/lottery_service ./microservice-backend/lottery
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/post_service ./microservice-backend/post
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/agent_service ./microservice-backend/agent
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/user_service ./microservice-backend/user
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/session_service ./microservice-backend/session
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/outbox_service ./backend/micro/outbox
	@GOOS=darwin GOARCH=arm64 go build -o ./app/macos/bff_service ./microservice-backend/bff
