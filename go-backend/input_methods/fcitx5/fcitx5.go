// Fcitx5 输入法 Go 实现
// 参考 rime 实现
package fcitx5

import (
	"log"
	"os"
	"path/filepath"

	"github.com/EasyIME/pime-go/pime"
)

// Fcitx5 类型定义
type Fcitx5Instance uintptr
type Fcitx5Context uintptr
type Fcitx5KeyCode uint32
type Fcitx5KeyState uint32

const (
	APP         = "PIME"
	APP_VERSION = "0.01"

	// 命令ID
	ID_MODE_ICON  = 2000
	ID_ASCII_MODE = 2001
	ID_FULL_SHAPE = 2002
	ID_SETTINGS   = 2003
)

// Style 样式配置
type Style struct {
	DisplayTrayIcon bool
}

// IME Fcitx5 输入法
type IME struct {
	*pime.TextServiceBase
	iconDir         string
	style           Style
	selectKeys      string
	lastKeyDownCode int
	lastKeySkip     int
	lastKeyDownRet  bool
	lastKeyUpCode   int
	lastKeyUpRet    bool
	keyComposing    bool
	// Fcitx5 相关
	instance Fcitx5Instance
	context  Fcitx5Context
}

// New 创建 Fcitx5 输入法实例
func New(client *pime.Client) pime.TextService {
	return &IME{
		TextServiceBase: pime.NewTextServiceBase(client),
		style: Style{
			DisplayTrayIcon: true,
		},
	}
}

// HandleRequest 处理请求
func (ime *IME) HandleRequest(req *pime.Request) *pime.Response {
	resp := pime.NewResponse(req.SeqNum, true)

	switch req.Method {
	case "onActivate":
		return ime.onActivate(req, resp)

	case "onDeactivate":
		return ime.onDeactivate(req, resp)

	case "filterKeyDown":
		return ime.filterKeyDown(req, resp)

	case "onKeyDown":
		return ime.onKeyDown(req, resp)

	case "filterKeyUp":
		resp.ReturnValue = 0

	case "onKeyUp":
		resp.ReturnValue = 0

	case "onCompositionTerminated":
		// 清理状态

	case "onCommand":
		return ime.onCommand(req, resp)

	default:
		resp.ReturnValue = 0
	}

	return resp
}

// onActivate 激活输入法
func (ime *IME) onActivate(req *pime.Request, resp *pime.Response) *pime.Response {
	// 获取图标目录
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		iconDir := filepath.Join(exeDir, "input_methods", "fcitx5", "icons")
		if dirExists(iconDir) {
			log.Println("Fcitx5 图标目录:", iconDir)
			// 添加托盘按钮
			if ime.style.DisplayTrayIcon {
				// 语言切换按钮
				iconPath := filepath.Join(iconDir, "eng.ico")
				if fileExists(iconPath) {
					resp.AddButton = append(resp.AddButton, pime.ButtonInfo{
						ID:        "switch-lang",
						Icon:      iconPath,
						Tooltip:   "中西文切换",
						CommandID: ID_ASCII_MODE,
					})
				} else {
					// 使用 rime 的图标作为备用
					iconPath = filepath.Join(exeDir, "input_methods", "rime", "icons", "eng.ico")
					if fileExists(iconPath) {
						resp.AddButton = append(resp.AddButton, pime.ButtonInfo{
							ID:        "switch-lang",
							Icon:      iconPath,
							Tooltip:   "中西文切换",
							CommandID: ID_ASCII_MODE,
						})
					}
				}

				// 全半角切换按钮
				iconPath = filepath.Join(iconDir, "full.ico")
				if fileExists(iconPath) {
					resp.AddButton = append(resp.AddButton, pime.ButtonInfo{
						ID:        "switch-shape",
						Icon:      iconPath,
						Tooltip:   "全角/半角切换",
						CommandID: ID_FULL_SHAPE,
					})
				} else {
					iconPath = filepath.Join(exeDir, "input_methods", "rime", "icons", "full.ico")
					if fileExists(iconPath) {
						resp.AddButton = append(resp.AddButton, pime.ButtonInfo{
							ID:        "switch-shape",
							Icon:      iconPath,
							Tooltip:   "全角/半角切换",
							CommandID: ID_FULL_SHAPE,
						})
					}
				}

				// 设置按钮
				iconPath = filepath.Join(iconDir, "config.ico")
				if fileExists(iconPath) {
					resp.AddButton = append(resp.AddButton, pime.ButtonInfo{
						ID:      "settings",
						Icon:    iconPath,
						Tooltip: "设置",
						Type:    "menu",
					})
				} else {
					iconPath = filepath.Join(exeDir, "input_methods", "rime", "icons", "config.ico")
					if fileExists(iconPath) {
						resp.AddButton = append(resp.AddButton, pime.ButtonInfo{
							ID:      "settings",
							Icon:    iconPath,
							Tooltip: "设置",
							Type:    "menu",
						})
					}
				}
			}
		}
	}

	log.Println("Fcitx5 输入法已激活")
	resp.ReturnValue = 1
	return resp
}

// onDeactivate 失活输入法
func (ime *IME) onDeactivate(req *pime.Request, resp *pime.Response) *pime.Response {
	log.Println("Fcitx5 输入法已失活")
	resp.ReturnValue = 1
	return resp
}

// filterKeyDown 过滤按键
func (ime *IME) filterKeyDown(req *pime.Request, resp *pime.Response) *pime.Response {
	return ime.onKeyDown(req, resp)
}

// onKeyDown 处理按键
func (ime *IME) onKeyDown(req *pime.Request, resp *pime.Response) *pime.Response {
	keyCode := req.KeyCode
	charCode := req.CharCode

	// 检查是否使用真实的 Fcitx5
	if ime.context != 0 {
		// 这里将来会实现真实的 Fcitx5 按键处理
		log.Println("使用真实 Fcitx5 处理按键")
	} else {
		// 使用模拟模式
		// 处理 'h' 键
		if charCode == 104 || charCode == 72 { // 'h' or 'H'
			resp.CompositionString = "ha"
			resp.CursorPos = 2
			resp.CandidateList = []string{"哈", "呵", "喝", "和", "河"}
			resp.ShowCandidates = true
			resp.ReturnValue = 1
			return resp
		}

		// 处理 'a' 键
		if charCode == 97 || charCode == 65 { // 'a' or 'A'
			if req.CompositionString == "ha" {
				resp.CompositionString = "ha"
				resp.CursorPos = 2
				resp.CandidateList = []string{"哈", "呵", "喝", "和", "河"}
				resp.ShowCandidates = true
				resp.ReturnValue = 1
				return resp
			}
		}

		// 处理数字键选择候选词
		if keyCode >= 0x31 && keyCode <= 0x35 { // '1' - '5'
			if len(req.CandidateList) > 0 {
				index := int(keyCode - 0x31)
				if index < len(req.CandidateList) {
					resp.CommitString = req.CandidateList[index]
					resp.ReturnValue = 1
					return resp
				}
			}
		}
	}

	// 其他按键不处理
	resp.ReturnValue = 0
	return resp
}

// onCommand 处理命令
func (ime *IME) onCommand(req *pime.Request, resp *pime.Response) *pime.Response {
	commandID, ok := req.Data["commandId"].(float64)
	if !ok {
		resp.ReturnValue = 0
		return resp
	}

	switch int(commandID) {
	case ID_MODE_ICON:
		log.Println("点击模式图标")

	case ID_ASCII_MODE:
		log.Println("切换中英文模式")

	case ID_FULL_SHAPE:
		log.Println("切换全半角模式")

	case ID_SETTINGS:
		log.Println("打开设置")

	default:
		log.Printf("未知命令: %d", int(commandID))
	}

	resp.ReturnValue = 1
	return resp
}

// Init 初始化
func (ime *IME) Init(req *pime.Request) bool {
	// 初始化 Fcitx5 环境
	log.Println("Fcitx5 输入法初始化")

	// 获取配置目录
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		configDir := filepath.Join(exeDir, "input_methods", "fcitx5", "data")

		// 创建配置目录
		if !dirExists(configDir) {
			err := os.MkdirAll(configDir, 0755)
			if err != nil {
				log.Printf("创建配置目录失败: %v", err)
			}
		}

		// 检查 fcitx5.dll 是否存在
		dllPath := filepath.Join(exeDir, "input_methods", "fcitx5", "fcitx5.dll")
		if fileExists(dllPath) {
			log.Println("找到 fcitx5.dll，准备使用真实 Fcitx5")
			// 这里将来会实现真实的 Fcitx5 初始化
		} else {
			log.Println("未找到 fcitx5.dll，使用模拟模式")
		}
	} else {
		log.Println("获取可执行文件路径失败，使用模拟模式")
	}

	return true
}

// Close 关闭
func (ime *IME) Close() {
	log.Println("Fcitx5 输入法关闭")

	// 清理 Fcitx5 资源
	// 这里将来会实现真实的 Fcitx5 资源清理
}

// 辅助函数
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
