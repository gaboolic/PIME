// RIME 输入法 Go 实现（简化版）
// 参考 python/rime_ime.py
package rime

import (
	"log"

	"github.com/EasyIME/pime-go/pime"
)

const (
	APP             = "PIME"
	APP_VERSION     = "0.01"
	CONFIG_FILE     = "PIME.yaml"

	// 命令ID
	ID_MODE_ICON    = 1000
	ID_ASCII_MODE   = 1001
	ID_FULL_SHAPE   = 1002
	ID_SETTINGS     = 1003
)

// Style 样式配置
type Style struct {
	DisplayTrayIcon bool
}

// IME RIME 输入法
type IME struct {
	*pime.TextServiceBase
	iconDir    string
	style      Style
	selectKeys string
	lastKeyDownCode int
	lastKeySkip     int
	lastKeyDownRet  bool
	lastKeyUpCode   int
	lastKeyUpRet    bool
	keyComposing    bool
	asciiMode       bool
}

func normalizeLetterCharCode(keyCode, charCode int) int {
	if charCode != 0 {
		return charCode
	}
	if keyCode >= 0x41 && keyCode <= 0x5A {
		return keyCode + 32
	}
	return charCode
}

// New 创建 RIME 输入法实例
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
	log.Println("RIME 输入法已激活")
	pime.AddLangButtons(resp, ime.Client, ime.asciiMode, ID_MODE_ICON, ID_ASCII_MODE)
	resp.ReturnValue = 1
	return resp
}

// onDeactivate 失活输入法
func (ime *IME) onDeactivate(req *pime.Request, resp *pime.Response) *pime.Response {
	log.Println("RIME 输入法已失活")
	pime.RemoveLangButtons(resp, ime.Client)
	resp.ReturnValue = 1
	return resp
}

// filterKeyDown 过滤按键
func (ime *IME) filterKeyDown(req *pime.Request, resp *pime.Response) *pime.Response {
	return ime.onKeyDown(req, resp)
}

// onKeyDown 处理按键
func (ime *IME) onKeyDown(req *pime.Request, resp *pime.Response) *pime.Response {
	// 简化实现：模拟 RIME 输入
	keyCode := req.KeyCode
	charCode := normalizeLetterCharCode(keyCode, req.CharCode)

	if ime.asciiMode && req.CompositionString == "" && len(req.CandidateList) == 0 && charCode >= 0x20 {
		resp.ReturnValue = 0
		return resp
	}

	// 处理 'n' 键
	if charCode == 110 || charCode == 78 { // 'n' or 'N'
		resp.CompositionString = "ni"
		resp.CursorPos = 2
		resp.CandidateList = []string{"你", "泥", "尼", "呢", "倪"}
		resp.ShowCandidates = true
		resp.ReturnValue = 1
		return resp
	}

	// 处理 'i' 键
	if charCode == 105 || charCode == 73 { // 'i' or 'I'
		if req.CompositionString == "ni" {
			resp.CompositionString = "ni"
			resp.CursorPos = 2
			resp.CandidateList = []string{"你", "泥", "尼", "呢", "倪"}
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
		ime.asciiMode = !ime.asciiMode
		pime.ChangeLangButtons(resp, ime.Client, ime.asciiMode)

	case ID_ASCII_MODE:
		ime.asciiMode = !ime.asciiMode
		pime.ChangeLangButtons(resp, ime.Client, ime.asciiMode)

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
	// 初始化 RIME 环境
	log.Println("RIME 输入法初始化")
	return true
}

// Close 关闭
func (ime *IME) Close() {
	log.Println("RIME 输入法关闭")
}

