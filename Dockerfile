FROM golang

WORKDIR /app
# go proxy for china
RUN go env -w GOPROXY=https://goproxy.cn

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]
