// PIME Go 后端主入口
// 参考 python/server.py 实现
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/EasyIME/pime-go/pime"

	// 导入输入法包
	"github.com/EasyIME/pime-go/input_methods/fcitx5"
	"github.com/EasyIME/pime-go/input_methods/meow"
	"github.com/EasyIME/pime-go/input_methods/rime"
	simplepinyin "github.com/EasyIME/pime-go/input_methods/simple_pinyin"
)

// Client 客户端连接
type Client struct {
	ID              string
	GUID            string
	IsWindows8Above bool
	IsMetroApp      bool
	IsUiLess        bool
	IsConsole       bool
	Service         pime.TextService
}

// ServiceFactory 服务工厂函数
type ServiceFactory func(clientID string, guid string) pime.TextService

// Server PIME 服务器
type Server struct {
	mu        sync.RWMutex
	clients   map[string]*Client
	factories map[string]ServiceFactory // guid -> factory
	reader    *bufio.Reader
	running   bool
}

// NewServer 创建服务器
func NewServer() *Server {
	return &Server{
		clients:   make(map[string]*Client),
		factories: make(map[string]ServiceFactory),
		reader:    bufio.NewReader(os.Stdin),
	}
}

// RegisterService 注册输入法服务工厂
func (s *Server) RegisterService(guid string, factory ServiceFactory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.factories[guid] = factory
	log.Printf("注册输入法服务: %s", guid)
}

// Run 运行服务器
func (s *Server) Run() error {
	s.running = true
	log.Println("PIME Go 后端服务器已启动")

	for s.running {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("收到 EOF，服务器停止")
				return nil
			}
			return fmt.Errorf("读取错误: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := s.handleMessage(line); err != nil {
			log.Printf("处理消息错误: %v", err)
			// 发送错误响应，防止客户端阻塞
			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				s.sendResponse(parts[0], map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
			}
		}
	}

	return nil
}

// handleMessage 处理消息
// 格式: "<client_id>|<JSON>"
func (s *Server) handleMessage(line string) error {
	parts := strings.SplitN(line, "|", 2)
	if len(parts) != 2 {
		return fmt.Errorf("无效的消息格式")
	}

	clientID := parts[0]
	jsonData := parts[1]

	var req pime.Request
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 处理请求
	resp := s.handleRequest(clientID, &req)

	// 发送响应
	return s.sendResponse(clientID, resp)
}

// handleRequest 处理请求
func (s *Server) handleRequest(clientID string, req *pime.Request) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch req.Method {
	case "init":
		// PIME 在 init 时通过顶层 id 传递语言配置 GUID。
		// 为了兼容已有调用，也接受 data.guid。
		guid := req.ID
		if guid == "" && req.Data != nil {
			guid, _ = req.Data["guid"].(string)
		}
		if guid == "" {
			return map[string]interface{}{
				"seqNum":  req.SeqNum,
				"success": false,
				"error":   "缺少 guid",
			}
		}

		// 创建客户端
		client := &Client{
			ID:              clientID,
			GUID:            guid,
			IsWindows8Above: req.IsWindows8Above,
			IsMetroApp:      req.IsMetroApp,
			IsUiLess:        req.IsUiLess,
			IsConsole:       req.IsConsole,
		}

		// 获取输入法服务工厂
		factory, ok := s.factories[guid]
		if !ok {
			return map[string]interface{}{
				"seqNum":  req.SeqNum,
				"success": false,
				"error":   fmt.Sprintf("未知的输入法: %s", guid),
			}
		}

		// 创建输入法服务
		client.Service = factory(clientID, guid)
		s.clients[clientID] = client

		// 初始化服务
		if !client.Service.Init(req) {
			delete(s.clients, clientID)
			return map[string]interface{}{
				"seqNum":  req.SeqNum,
				"success": false,
				"error":   "初始化失败",
			}
		}

		return map[string]interface{}{
			"seqNum":  req.SeqNum,
			"success": true,
		}

	case "onActivate", "onDeactivate", "filterKeyDown", "onKeyDown",
		"filterKeyUp", "onKeyUp", "onCommand", "onCompositionTerminated",
		"onPreservedKey", "onLangProfileActivated":
		// 转发到输入法服务
		client, ok := s.clients[clientID]
		if !ok {
			return map[string]interface{}{
				"seqNum":  req.SeqNum,
				"success": false,
				"error":   "客户端未初始化",
			}
		}

		resp := client.Service.HandleRequest(req)
		return s.convertResponse(resp)

	default:
		return map[string]interface{}{
			"seqNum":  req.SeqNum,
			"success": false,
			"error":   fmt.Sprintf("未知的方法: %s", req.Method),
		}
	}
}

// sendResponse 发送响应
func (s *Server) sendResponse(clientID string, resp map[string]interface{}) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("序列化响应失败: %w", err)
	}

	fmt.Printf("%s|%s|%s\n", pime.MsgPIME, clientID, string(data))
	return nil
}

// convertResponse 转换响应格式
func (s *Server) convertResponse(resp *pime.Response) map[string]interface{} {
	m := map[string]interface{}{
		"success":        resp.Success,
		"seqNum":         resp.SeqNum,
		"return":         resp.ReturnValue,
		"cursorPos":      resp.CursorPos,
		"showCandidates": resp.ShowCandidates,
	}

	if len(resp.CandidateList) > 0 {
		m["candidateList"] = resp.CandidateList
	}
	if resp.CompositionString != "" {
		m["compositionString"] = resp.CompositionString
	}
	if resp.CommitString != "" {
		m["commitString"] = resp.CommitString
	}
	if resp.SelStart != 0 || resp.SelEnd != 0 {
		m["selStart"] = resp.SelStart
		m["selEnd"] = resp.SelEnd
	}
	if len(resp.AddButton) > 0 {
		m["addButton"] = resp.AddButton
	}
	if len(resp.RemoveButton) > 0 {
		m["removeButton"] = resp.RemoveButton
	}
	if len(resp.ChangeButton) > 0 {
		m["changeButton"] = resp.ChangeButton
	}
	return m
}

// loadInputMethods 加载所有输入法
func loadInputMethods(server *Server) {
	// 获取当前目录
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("获取可执行文件路径失败:", err)
	}
	exeDir := filepath.Dir(exePath)

	// 扫描 input_methods 目录
	inputMethodsDir := filepath.Join(exeDir, "input_methods")
	entries, err := os.ReadDir(inputMethodsDir)
	if err != nil {
		log.Printf("读取 input_methods 目录失败: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 读取 ime.json
		imePath := filepath.Join(inputMethodsDir, entry.Name(), "ime.json")
		data, err := os.ReadFile(imePath)
		if err != nil {
			log.Printf("读取 %s 失败: %v", imePath, err)
			continue
		}

		var imeConfig map[string]interface{}
		if err := json.Unmarshal(data, &imeConfig); err != nil {
			log.Printf("解析 %s 失败: %v", imePath, err)
			continue
		}

		guid, _ := imeConfig["guid"].(string)
		name, _ := imeConfig["name"].(string)
		if guid == "" {
			log.Printf("%s 缺少 guid", entry.Name())
			continue
		}

		log.Printf("加载输入法: %s (%s)", name, guid)

		// 根据输入法名称注册不同的服务实现
		switch entry.Name() {
		case "meow":
			// 喵喵输入法
			server.RegisterService(guid, func(clientID, g string) pime.TextService {
				client := &pime.Client{ID: clientID}
				return meow.New(client)
			})
		case "rime":
			// RIME 输入法
			server.RegisterService(guid, func(clientID, g string) pime.TextService {
				client := &pime.Client{ID: clientID}
				return rime.New(client)
			})
		case "simple_pinyin":
			// 拼音输入法
			server.RegisterService(guid, func(clientID, g string) pime.TextService {
				client := &pime.Client{ID: clientID}
				return simplepinyin.New(client)
			})
		case "fcitx5":
			// Fcitx5 输入法
			server.RegisterService(guid, func(clientID, g string) pime.TextService {
				client := &pime.Client{ID: clientID}
				return fcitx5.New(client)
			})
		default:
			// 默认使用拼音输入法
			server.RegisterService(guid, func(clientID, g string) pime.TextService {
				client := &pime.Client{ID: clientID}
				return simplepinyin.New(client)
			})
		}
	}
}

func main() {
	// 设置日志
	logFile, err := os.OpenFile("go_backend.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("=" + strings.Repeat("=", 50))
	log.Println("PIME Go 后端启动")
	log.Println("=" + strings.Repeat("=", 50))

	// 创建服务器
	server := NewServer()

	// 加载所有输入法
	loadInputMethods(server)

	// 运行服务器
	if err := server.Run(); err != nil {
		log.Fatal("服务器错误:", err)
	}
}
