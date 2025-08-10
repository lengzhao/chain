package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/govm-net/chain/chain"
	"github.com/govm-net/chain/config"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建区块链实例
	blockchain, err := chain.New(cfg)
	if err != nil {
		log.Fatalf("创建区块链实例失败: %v", err)
	}

	// 启动区块链
	if err := blockchain.Start(); err != nil {
		log.Fatalf("启动区块链失败: %v", err)
	}

	fmt.Println("区块链系统已启动")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("正在关闭区块链系统...")
	blockchain.Stop()
	fmt.Println("区块链系统已关闭")
}
