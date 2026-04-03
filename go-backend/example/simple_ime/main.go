// 简单的 Go PIME 输入法后端示例
package main

import (
	"flag"
	"log"
	"os"

	"github.com/EasyIME/pime-go/pime"
)

func main() {
	// 解析命令行参数
	var name string
	flag.StringVar(&name, "n", "SimpleIME", "输入法名称")
	flag.Parse()

	// 设置日志输出到文件
	logFile, err := os.OpenFile("go_ime.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
	defer logFile.Close()

	// 同时输出到文件和stderr
	log.SetOutput(logFile)
	log.SetPrefix("[" + name + "] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("启动 PIME Go 后端:", name)
	log.Println("工作目录:", getWorkDir())

	// 创建服务管理器
	mgr := pime.NewServiceManager()

	// 注册输入法工厂
	mgr.Register("simple", func(clientID string) pime.TextService {
		log.Println("创建新客户端:", clientID)
		client := &pime.Client{ID: clientID}
		return NewSimpleIME(client)
	})

	// 启动服务
	log.Println("服务已启动，等待连接...")
	if err := mgr.Run(); err != nil {
		log.Fatal("服务错误:", err)
	}
}

func getWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "未知"
	}
	return dir
}
