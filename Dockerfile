# docker build -t ingress-manager:v0.0.1 .

FROM golang:1.19 as builder
WORKDIR /app
COPY . .
#ENV GOPROXY https://goproxy.cn
RUN GOPROXY=https://goproxy.cn  go mod download && CGO_ENABLED=0 go build -o ingress-manager ingressmanager/main.go


# binfile
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/ingress-manager .
CMD ["./ingress-manager"]
