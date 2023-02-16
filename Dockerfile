FROM golang as builder
WORKDIR /src/service
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=`git tag --sort=-version:refname | head -n 1`" -o rms-bot-server -a -installsuffix cgo rms-bot-server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN mkdir /app
WORKDIR /app
COPY --from=builder /src/service/rms-bot-server .
COPY --from=builder /src/service/configs/rms-bot-server.json /etc/rms/
CMD ["./rms-bot-server"]