package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pte_live_im/define"
	"pte_live_im/grpcserver"
	"pte_live_im/pkg/etcd"
	"pte_live_im/pkg/pulsar"
	"pte_live_im/pkg/redis"
	"pte_live_im/pkg/setting"
	"pte_live_im/queue"
	"pte_live_im/routers"
	"pte_live_im/servers"
	"pte_live_im/tools/log"
	"pte_live_im/tools/util"
)

func init() {
	setting.Setup()
	log.Setup()
}

func main() {
	initGRPCServer()

	if setting.RedisSetting.Addr != "" {
		if err := redis.Ping(redisCtx()); err != nil {
			fmt.Printf("Redis 连接失败（%s）: %v，直播房间状态将使用内存兜底\n", setting.RedisSetting.Addr, err)
		} else {
			fmt.Printf("Redis 已连接：%s\n", setting.RedisSetting.Addr)
		}
	}

	if setting.QueueSetting.Backend == "pulsar" || setting.QueueSetting.Backend == "both" {
		if err := pulsar.Setup(); err != nil {
			fmt.Printf("Pulsar 连接失败: %v\n", err)
			if setting.QueueSetting.Backend == "pulsar" {
				fmt.Println("queue.backend=pulsar 时 Pulsar 不可用，消息将同步 dispatch 或入队失败")
			}
		}
	}

	queue.StartWorkers()
	queue.StartChatWorkers()

	registerServer()
	routers.Init()
	servers.PingTimer()

	addr := ":" + setting.CommonSetting.HttpPort
	srv := &http.Server{Addr: addr, Handler: http.DefaultServeMux}

	go func() {
		fmt.Printf("HTTP API 启动，端口号：%s\n", setting.CommonSetting.HttpPort)
		fmt.Printf("WebSocket 启动，端口号：%s\n", setting.CommonSetting.WebSocketPort)
		fmt.Printf("gRPC API 启动，端口号：%s\n", setting.CommonSetting.RPCPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func initGRPCServer() {
	grpcserver.Init()
}

func registerServer() {
	if util.IsCluster() {
		ser, err := etcd.NewServiceReg(setting.EtcdSetting.Endpoints, 5)
		if err != nil {
			panic(err)
		}

		hostPort := net.JoinHostPort(setting.GlobalSetting.LocalHost, setting.CommonSetting.RPCPort)
		err = ser.PutService(define.ETCD_SERVER_LIST+hostPort, hostPort)
		if err != nil {
			panic(err)
		}

		cli, err := etcd.NewClientDis(setting.EtcdSetting.Endpoints)
		if err != nil {
			panic(err)
		}
		_, err = cli.GetService(define.ETCD_SERVER_LIST)
		if err != nil {
			panic(err)
		}
	}
}

func redisCtx() context.Context {
	return context.Background()
}
