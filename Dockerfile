FROM golang:alpine
WORKDIR /app
COPY . /app
RUN go build -o zyzzyva cmd/zyzzyva/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=0 /app/zyzzyva /app/zyzzyva
ENTRYPOINT ["./zyzzyva"]
