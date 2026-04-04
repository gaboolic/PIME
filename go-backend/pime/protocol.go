// PIME 通信协议定义
package pime

import (
	"encoding/json"
	"fmt"
)

// 消息类型
const (
	MsgPIME = "PIME_MSG"
)

type KeyStates []bool

func (k *KeyStates) UnmarshalJSON(data []byte) error {
	var bools []bool
	if err := json.Unmarshal(data, &bools); err == nil {
		*k = KeyStates(bools)
		return nil
	}

	var ints []int
	if err := json.Unmarshal(data, &ints); err == nil {
		states := make(KeyStates, len(ints))
		for i, v := range ints {
			states[i] = v != 0
		}
		*k = states
		return nil
	}

	return fmt.Errorf("invalid keyStates payload: %s", string(data))
}

// Request PIME请求结构
type Request struct {
	Method        string                 `json:"method"`
	SeqNum        int                    `json:"seqNum"`
	ID            string                 `json:"id,omitempty"`
	IsWindows8Above bool               `json:"isWindows8Above,omitempty"`
	IsMetroApp    bool                   `json:"isMetroApp,omitempty"`
	IsUiLess      bool                   `json:"isUiLess,omitempty"`
	IsConsole     bool                   `json:"isConsole,omitempty"`
	CharCode      int                    `json:"charCode,omitempty"`
	KeyCode       int                    `json:"keyCode,omitempty"`
	RepeatCount   int                    `json:"repeatCount,omitempty"`
	ScanCode      int                    `json:"scanCode,omitempty"`
	IsExtended    bool                   `json:"isExtended,omitempty"`
	KeyStates     KeyStates              `json:"keyStates,omitempty"`
	CompositionString string             `json:"compositionString,omitempty"`
	CandidateList []string               `json:"candidateList,omitempty"`
	ShowCandidates bool                  `json:"showCandidates,omitempty"`
	CursorPos     int                    `json:"cursorPos,omitempty"`
	SelStart      int                    `json:"selStart,omitempty"`
	SelEnd        int                    `json:"selEnd,omitempty"`
	// 扩展字段
	Data          map[string]interface{} `json:"data,omitempty"`
}

// ButtonInfo 按钮信息
type ButtonInfo struct {
	ID        string `json:"id"`
	Icon      string `json:"icon,omitempty"`
	Tooltip   string `json:"tooltip,omitempty"`
	CommandID int    `json:"commandId,omitempty"`
	Type      string `json:"type,omitempty"` // "button", "toggle", "menu"
	Enable    bool   `json:"enable,omitempty"`
	Toggled   bool   `json:"toggled,omitempty"`
}

// Response PIME响应结构
type Response struct {
	SeqNum           int          `json:"seqNum"`
	Success          bool         `json:"success"`
	ReturnValue      int          `json:"returnValue,omitempty"`
	CompositionString string       `json:"compositionString,omitempty"`
	CommitString     string       `json:"commitString,omitempty"`
	CandidateList    []string     `json:"candidateList,omitempty"`
	ShowCandidates   bool         `json:"showCandidates,omitempty"`
	CursorPos        int          `json:"cursorPos,omitempty"`
	SelStart         int          `json:"selStart,omitempty"`
	SelEnd           int          `json:"selEnd,omitempty"`
	Message          string       `json:"message,omitempty"`
	// 按钮相关
	AddButton    []ButtonInfo `json:"addButton,omitempty"`
	RemoveButton []string     `json:"removeButton,omitempty"`
	ChangeButton []ButtonInfo `json:"changeButton,omitempty"`
}

// ParseRequest 解析请求消息
func ParseRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("unmarshal request failed: %w", err)
	}
	return &req, nil
}

// ToJSON 转换为 JSON 字节
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// NewResponse 创建新响应
func NewResponse(seqNum int, success bool) *Response {
	return &Response{
		SeqNum:  seqNum,
		Success: success,
	}
}
