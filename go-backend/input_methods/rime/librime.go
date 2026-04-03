// RIME C API 封装
// 参考 python/librime.py
package rime

/*
#cgo LDFLAGS: -L${SRCDIR} -lrime
#include <stdint.h>
#include <stdlib.h>

// RIME 结构体定义
typedef void* RimeSessionId;
typedef int Bool;

typedef struct {
    int data_size;
    const char* shared_data_dir;
    const char* user_data_dir;
    const char* distribution_name;
    const char* distribution_code_name;
    const char* distribution_version;
    const char* app_name;
    const char** modules;
} RimeTraits;

typedef struct {
    int length;
    int cursor_pos;
    int sel_start;
    int sel_end;
    const char* preedit;
} RimeComposition;

typedef struct {
    const char* text;
    const char* comment;
    void* reserved;
} RimeCandidate;

typedef struct {
    int page_size;
    int page_no;
    Bool is_last_page;
    int highlighted_candidate_index;
    int num_candidates;
    RimeCandidate* candidates;
    const char* select_keys;
} RimeMenu;

typedef struct {
    int data_size;
    const char* text;
} RimeCommit;

// RIME 函数声明
extern Bool rime_init(const RimeTraits* traits);
extern void rime_finalize(void);
extern Bool rime_start_session(RimeSessionId* session_id);
extern void rime_end_session(RimeSessionId session_id);
extern Bool rime_process_key(RimeSessionId session_id, int keycode, int modifiers);
extern Bool rime_get_composition(RimeSessionId session_id, RimeComposition* composition);
extern Bool rime_get_menu(RimeSessionId session_id, RimeMenu* menu);
extern Bool rime_get_commit(RimeSessionId session_id, RimeCommit* commit);
extern void rime_free_commit(RimeCommit* commit);
extern void rime_select_candidate(RimeSessionId session_id, int index);
extern void rime_select_page(RimeSessionId session_id, int page_no);
extern Bool rime_deploy_config_file(const char* file_path, const char* key);
extern void rime_set_notification_handler(void* context, void (*handler)(void* context, RimeSessionId session_id, const char* message_type, const char* message_value));

extern const char* rime_api_version(void);
extern const char* rime_get_name(void);
extern const char* rime_get_version(void);

// 辅助函数
static inline void free_strings(const char** strings) {
    if (strings) {
        for (int i = 0; strings[i]; i++) {
            free((void*)strings[i]);
        }
        free(strings);
    }
}

static inline char** make_c_str_array(const char* const* strings) {
    int count = 0;
    while (strings && strings[count]) {
        count++;
    }
    char** c_strings = (char**)malloc((count + 1) * sizeof(char*));
    if (c_strings) {
        for (int i = 0; i < count; i++) {
            c_strings[i] = (char*)strings[i];
        }
        c_strings[count] = NULL;
    }
    return c_strings;
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

const (
	RIME_MAX_NUM_CANDIDATES = 10
)

// RimeSessionId RIME 会话ID
type RimeSessionId C.RimeSessionId

// RimeTraits RIME 初始化参数
type RimeTraits struct {
	SharedDataDir        string
	UserDataDir          string
	DistributionName     string
	DistributionCodeName string
	DistributionVersion  string
	AppName              string
	Modules              []string
}

// RimeComposition 组合字符串
type RimeComposition struct {
	Length    int
	CursorPos int
	SelStart  int
	SelEnd    int
	Preedit   string
}

// RimeCandidate 候选词
type RimeCandidate struct {
	Text    string
	Comment string
}

// RimeMenu 候选词菜单
type RimeMenu struct {
	PageSize                int
	PageNo                  int
	IsLastPage              bool
	HighlightedCandidateIndex int
	NumCandidates           int
	Candidates              []RimeCandidate
	SelectKeys              string
}

// RimeCommit 提交文本
type RimeCommit struct {
	Text string
}

// NotificationHandler 通知回调函数类型
type NotificationHandler func(session RimeSessionId, messageType, messageValue string)

// notificationHandler C 回调函数
func notificationHandler(context unsafe.Pointer, session C.RimeSessionId, messageType, messageValue *C.char) {
	if context == nil {
		return
	}
	handler := (*NotificationHandler)(context)
	if handler != nil {
		*handler(RimeSessionId(session), C.GoString(messageType), C.GoString(messageValue))
	}
}

// Init 初始化 RIME
func Init(traits RimeTraits) bool {
	var cTraits C.RimeTraits
	cTraits.data_size = C.int(unsafe.Sizeof(cTraits))
	cTraits.shared_data_dir = C.CString(traits.SharedDataDir)
	defer C.free(unsafe.Pointer(cTraits.shared_data_dir))
	cTraits.user_data_dir = C.CString(traits.UserDataDir)
	defer C.free(unsafe.Pointer(cTraits.user_data_dir))
	cTraits.distribution_name = C.CString(traits.DistributionName)
	defer C.free(unsafe.Pointer(cTraits.distribution_name))
	cTraits.distribution_code_name = C.CString(traits.DistributionCodeName)
	defer C.free(unsafe.Pointer(cTraits.distribution_code_name))
	cTraits.distribution_version = C.CString(traits.DistributionVersion)
	defer C.free(unsafe.Pointer(cTraits.distribution_version))
	cTraits.app_name = C.CString(traits.AppName)
	defer C.free(unsafe.Pointer(cTraits.app_name))

	if len(traits.Modules) > 0 {
		cModules := make([]*C.char, len(traits.Modules)+1)
		for i, module := range traits.Modules {
			cModules[i] = C.CString(module)
			defer C.free(unsafe.Pointer(cModules[i]))
		}
		cModules[len(traits.Modules)] = nil
		cTraits.modules = (**C.char)(unsafe.Pointer(&cModules[0]))
	}

	return bool(C.rime_init(&cTraits))
}

// Finalize 清理 RIME
func Finalize() {
	C.rime_finalize()
}

// StartSession 开始会话
func StartSession() (RimeSessionId, bool) {
	var sessionId RimeSessionId
	success := bool(C.rime_start_session((*C.RimeSessionId)(&sessionId)))
	return sessionId, success
}

// EndSession 结束会话
func EndSession(sessionId RimeSessionId) {
	C.rime_end_session(C.RimeSessionId(sessionId))
}

// ProcessKey 处理按键
func ProcessKey(sessionId RimeSessionId, keyCode, modifiers int) bool {
	return bool(C.rime_process_key(C.RimeSessionId(sessionId), C.int(keyCode), C.int(modifiers)))
}

// GetComposition 获取组合字符串
func GetComposition(sessionId RimeSessionId) (RimeComposition, bool) {
	var cComposition C.RimeComposition
	success := bool(C.rime_get_composition(C.RimeSessionId(sessionId), &cComposition))
	if !success {
		return RimeComposition{}, false
	}

	composition := RimeComposition{
		Length:    int(cComposition.length),
		CursorPos: int(cComposition.cursor_pos),
		SelStart:  int(cComposition.sel_start),
		SelEnd:    int(cComposition.sel_end),
		Preedit:   C.GoString(cComposition.preedit),
	}
	return composition, true
}

// GetMenu 获取候选词菜单
func GetMenu(sessionId RimeSessionId) (RimeMenu, bool) {
	var cMenu C.RimeMenu
	success := bool(C.rime_get_menu(C.RimeSessionId(sessionId), &cMenu))
	if !success {
		return RimeMenu{}, false
	}

	menu := RimeMenu{
		PageSize:                int(cMenu.page_size),
		PageNo:                  int(cMenu.page_no),
		IsLastPage:              bool(cMenu.is_last_page),
		HighlightedCandidateIndex: int(cMenu.highlighted_candidate_index),
		NumCandidates:           int(cMenu.num_candidates),
		SelectKeys:              C.GoString(cMenu.select_keys),
	}

	if cMenu.num_candidates > 0 {
		menu.Candidates = make([]RimeCandidate, 0, cMenu.num_candidates)
		candidates := (*[1 << 20]C.RimeCandidate)(unsafe.Pointer(cMenu.candidates))[:cMenu.num_candidates:cMenu.num_candidates]
		for _, cCandidate := range candidates {
			candidate := RimeCandidate{
				Text:    C.GoString(cCandidate.text),
				Comment: C.GoString(cCandidate.comment),
			}
			menu.Candidates = append(menu.Candidates, candidate)
		}
	}

	return menu, true
}

// GetCommit 获取提交文本
func GetCommit(sessionId RimeSessionId) (RimeCommit, bool) {
	var cCommit C.RimeCommit
	success := bool(C.rime_get_commit(C.RimeSessionId(sessionId), &cCommit))
	if !success {
		return RimeCommit{}, false
	}
	defer C.rime_free_commit(&cCommit)

	commit := RimeCommit{
		Text: C.GoString(cCommit.text),
	}
	return commit, true
}

// SelectCandidate 选择候选词
func SelectCandidate(sessionId RimeSessionId, index int) {
	C.rime_select_candidate(C.RimeSessionId(sessionId), C.int(index))
}

// SelectPage 选择页码
func SelectPage(sessionId RimeSessionId, pageNo int) {
	C.rime_select_page(C.RimeSessionId(sessionId), C.int(pageNo))
}

// DeployConfigFile 部署配置文件
func DeployConfigFile(filePath, key string) bool {
	cFilePath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cFilePath))
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	return bool(C.rime_deploy_config_file(cFilePath, cKey))
}

// SetNotificationHandler 设置通知回调
func SetNotificationHandler(handler NotificationHandler) {
	// 暂时不支持通知回调，因为 Go 不允许在非 main 包中使用 //export
	C.rime_set_notification_handler(nil, nil)
}

// API 版本相关函数
func APIVersion() string {
	return C.GoString(C.rime_api_version())
}

func GetName() string {
	return C.GoString(C.rime_get_name())
}

func GetVersion() string {
	return C.GoString(C.rime_get_version())
}

// RimeInit 初始化 RIME 环境
func RimeInit(datadir, userdir, appname, appver string, fullcheck bool) bool {
	// 确保用户目录存在
	if err := os.MkdirAll(userdir, 0700); err != nil {
		fmt.Printf("创建用户目录失败: %v\n", err)
		return false
	}

	traits := RimeTraits{
		SharedDataDir:        datadir,
		UserDataDir:          userdir,
		DistributionName:     "PIME",
		DistributionCodeName: "pime",
		DistributionVersion:  appver,
		AppName:              appname,
	}

	if !Init(traits) {
		fmt.Println("RIME 初始化失败")
		return false
	}

	// 部署配置文件
	configFile := filepath.Join(datadir, "PIME.yaml")
	if !DeployConfigFile(configFile, "config_version") {
		fmt.Println("部署配置文件失败")
		return false
	}

	// 设置通知回调
	SetNotificationHandler(func(session RimeSessionId, messageType, messageValue string) {
		// 处理通知
		// fmt.Printf("RIME 通知: %s - %s\n", messageType, messageValue)
	})

	return true
}
