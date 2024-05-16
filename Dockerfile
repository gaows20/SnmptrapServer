# 使用官方Golang镜像作为基础镜像
FROM golang:1.18.2

# 设置工作目录
WORKDIR /go/src/app

# 复制项目文件到容器中
COPY . .

# 编译项目
RUN go get
RUN go build -tags gosnmp_nodebug -o app main.go 

EXPOSE 162/udp
EXPOSE 8070/tcp

# 运行应用程序
CMD ["./app"]
