# 基础镜像
FROM ubuntu:24.04

# 指定作者
LABEL authors="yzletter"

# 把本地编译后的程序拷贝到目标目录
COPY go-postery /app/go-postery

# 指定工作路径
WORKDIR /app

# 启动（最佳实践, CMD 是执行一个命令）
ENTRYPOINT ["/app/go-postery"]