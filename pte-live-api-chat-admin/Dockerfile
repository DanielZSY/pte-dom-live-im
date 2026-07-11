FROM golang:1.24-alpine AS builder

ARG APK_MIRROR=dl-cdn.alpinelinux.org
WORKDIR /src
RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -o /out/pte_live_api_chat_admin ./cmd/server

FROM alpine:3.20
ARG APK_MIRROR=dl-cdn.alpinelinux.org
RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata wget
WORKDIR /data/pte_live_api_chat_admin
COPY --from=builder /out/pte_live_api_chat_admin ./pte_live_api_chat_admin
COPY conf/app.yaml.example ./conf/app.yaml.example
EXPOSE 11505
CMD ["./pte_live_api_chat_admin", "-c", "./conf/app.yaml"]
