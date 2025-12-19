package shortcut

import (
	"Q-Solver/pkg/logger"
	"Q-Solver/pkg/winapi"
	"fmt"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

type KeyBinding struct {
	ComboID string `json:"vkCode"`
	KeyName string `json:"keyName"`
}

type Manager struct {
	hHook           uintptr
	hMouseHook      uintptr
	recordingKeyFor string
	maxComboKeys    map[uint32]bool
	heldKeys        map[uint32]bool

	// Callbacks
	OnTrigger           func(action string)
	OnRecord            func(action string, keyName string, comboID string)
	OnRecordingComplete func(action string, keyName string, comboID string)
	OnError             func(msg string)

	// Configuration
	Shortcuts map[string]KeyBinding
}

var globalManager *Manager

func NewManager() *Manager {
	return &Manager{
		heldKeys:     make(map[uint32]bool),
		maxComboKeys: make(map[uint32]bool),
		Shortcuts:    make(map[string]KeyBinding),
	}
}

func (m *Manager) Start() {
	globalManager = m
	go m.installHooks()
}

func (m *Manager) Stop() {
	if m.hHook != 0 {
		if winapi.UnhookWindowsHookEx(m.hHook) {
			logger.Println("卸载键盘Hook成功")
		} else {
			logger.Println("卸载键盘Hook失败")
		}
		m.hHook = 0
	}
	if m.hMouseHook != 0 {
		if winapi.UnhookWindowsHookEx(m.hMouseHook) {
			logger.Println("卸载鼠标Hook成功")
		} else {
			logger.Println("卸载鼠标Hook失败")
		}
		m.hMouseHook = 0
	}
	globalManager = nil
}

func (m *Manager) StartRecording(action string) {
	m.recordingKeyFor = action
	m.maxComboKeys = make(map[uint32]bool)
	logger.Printf("开始录制快捷键: %s\n", action)
}

func (m *Manager) StopRecording() {
	m.recordingKeyFor = ""
	logger.Println("停止录制快捷键")
}

func (m *Manager) installHooks() {
	// 获取模块句柄
	hMod := winapi.GetModuleHandle("")

	// 创建键盘回调
	kbdCallback := syscall.NewCallback(keyboardHookProc)
	// 安装键盘钩子
	m.hHook = winapi.SetWindowsHookEx(winapi.WH_KEYBOARD_LL, kbdCallback, hMod, 0)
	if m.hHook == 0 {
		logger.Println("安装键盘钩子失败")
	}

	// 创建鼠标回调
	mouseCallback := syscall.NewCallback(mouseHookProc)
	// 安装鼠标钩子
	m.hMouseHook = winapi.SetWindowsHookEx(winapi.WH_MOUSE_LL, mouseCallback, hMod, 0)
	if m.hMouseHook == 0 {
		logger.Println("安装鼠标钩子失败")
	}

	if m.hHook == 0 && m.hMouseHook == 0 {
		return
	}

	// 消息循环
	var msg winapi.MSG
	for winapi.GetMessage(&msg, 0, 0, 0) > 0 {
		// 保持线程活跃以处理钩子消息
	}
}

//这里解释了为什么只能吞掉第二个键，所以导致丢失焦点的问题：（其实是因为alt键的问题）
// 第一个键（Alt）按下：

// 记录：heldKeys = {Alt}
// 判断：有快捷键是只按 Alt 的吗？ -> 没有。
// 结果：放行（Chrome 收到 Alt）。
// 第二个键（~）按下：

// 记录：heldKeys = {Alt, ~}
// 判断：有快捷键是 Alt + ~ 的吗？ -> 有！
// 结果：拦截（return 1，Chrome 收不到 ~）。
func keyboardHookProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if globalManager == nil {
		return 0
	}
	// 只有当 nCode >= 0 时才处理消息，否则直接放行
	if nCode >= 0 {
		// 将 lParam 指针转换为键盘钩子结构体
		kbd := (*winapi.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		// 监听按下事件 (WM_KEYDOWN) 或 系统按键按下 (WM_SYSKEYDOWN，比如按住 Alt 时)
		if wParam == winapi.WM_KEYDOWN || wParam == winapi.WM_SYSKEYDOWN {
			globalManager.heldKeys[kbd.VkCode] = true
			if onKeysChanged() {
				return 1
			}
		}
		// 处理松开事件
		if wParam == winapi.WM_KEYUP || wParam == winapi.WM_SYSKEYUP {
			// 1. 从 map 中移除该键
			delete(globalManager.heldKeys, kbd.VkCode)

			// 录制模式下，松开按键也要检查是否结束录制
			if globalManager.recordingKeyFor != "" {
				if len(globalManager.heldKeys) == 0 {
					finishRecording()
				}
				return 1 // 录制期间吞掉所有按键
			}
		}
	}

	// 如果不是我们要拦截的键，或者 nCode < 0，必须调用 CallNextHookEx
	// 否则会导致系统键盘卡死或其他人无法使用键盘
	return winapi.CallNextHookEx(globalManager.hHook, nCode, wParam, lParam)
}

func mouseHookProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if globalManager == nil {
		return 0
	}
	if nCode >= 0 {
		mouseStruct := (*winapi.MSLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		var vkCode uint32
		isDown := false
		isUp := false

		switch wParam {
		case winapi.WM_XBUTTONDOWN:
			isDown = true
			xButton := (mouseStruct.MouseData >> 16) & 0xFFFF
			switch xButton {
			case 1:
				vkCode = winapi.VK_XBUTTON1
			case 2:
				vkCode = winapi.VK_XBUTTON2
			}
		case winapi.WM_XBUTTONUP:
			isUp = true
			xButton := (mouseStruct.MouseData >> 16) & 0xFFFF
			switch xButton {
			case 1:
				vkCode = winapi.VK_XBUTTON1
			case 2:
				vkCode = winapi.VK_XBUTTON2
			}
		}

		if vkCode != 0 {
			if isDown {
				globalManager.heldKeys[vkCode] = true
				if onKeysChanged() {
					return 1
				}
			} else if isUp {
				delete(globalManager.heldKeys, vkCode)
				// 录制模式下，松开按键也要检查是否结束录制
				if globalManager.recordingKeyFor != "" {
					if len(globalManager.heldKeys) == 0 {
						finishRecording()
					}
					return 1
				}
			}
		}
	}
	return winapi.CallNextHookEx(globalManager.hMouseHook, nCode, wParam, lParam)
}

func onKeysChanged() bool {
	if globalManager == nil {
		return false
	}

	// --- 录制模式 ---
	if globalManager.recordingKeyFor != "" {
		// 更新最大按键组合
		if len(globalManager.heldKeys) >= len(globalManager.maxComboKeys) {
			globalManager.maxComboKeys = make(map[uint32]bool)
			for k, v := range globalManager.heldKeys {
				globalManager.maxComboKeys[k] = v
			}
		}

		// 实时发给前端显示
		readableName := GetReadableName(globalManager.maxComboKeys)
		if globalManager.OnRecord != nil {
			globalManager.OnRecord(globalManager.recordingKeyFor, readableName, GetComboID(globalManager.maxComboKeys))
		}
		return true // 吞掉按键
	}

	// --- 正常模式 ---
	// 将当前按下的所有键生成 ID，去配置里查
	currentComboID := GetComboID(globalManager.heldKeys)
	for action, savedComboID := range globalManager.Shortcuts {
		if savedComboID.ComboID == currentComboID {
			// 检查是否包含 Alt 键 (VK_MENU=18, VK_LMENU=164, VK_RMENU=165)
			// 或者 Win 键 (VK_LWIN=91, VK_RWIN=92)
			// 如果包含这些键，且其他键被我们吞掉了，Windows 会认为用户只按了 Alt/Win，从而激活菜单栏或开始菜单
			hasAlt := globalManager.heldKeys[18] || globalManager.heldKeys[164] || globalManager.heldKeys[165]
			hasWin := globalManager.heldKeys[91] || globalManager.heldKeys[92]

			if hasAlt || hasWin {
				// 模拟按下并松开 Ctrl 键，防止 Windows 激活菜单栏/开始菜单
				// VK_CONTROL = 0x11, KEYEVENTF_KEYUP = 0x0002
				winapi.KeybdEvent(winapi.VK_CONTROL, 0, 0, 0)
				winapi.KeybdEvent(winapi.VK_CONTROL, 0, 2, 0)
			}

			if globalManager.OnTrigger != nil {
				globalManager.OnTrigger(action)
			}
			return true // 吞掉按键
		}
	}
	return false
}

func finishRecording() {
	if globalManager == nil || globalManager.recordingKeyFor == "" {
		return
	}

	// 如果没有按任何键（比如直接点击录制然后点别的），忽略
	if len(globalManager.maxComboKeys) == 0 {
		globalManager.recordingKeyFor = ""
		return
	}

	comboID := GetComboID(globalManager.maxComboKeys)
	readableName := GetReadableName(globalManager.maxComboKeys)
	action := globalManager.recordingKeyFor

	// 退出录制模式
	globalManager.recordingKeyFor = ""
	globalManager.maxComboKeys = nil

	// 异步调用回调，避免阻塞 Hook 线程
	go func() {
		if globalManager.OnRecordingComplete != nil {
			globalManager.OnRecordingComplete(action, readableName, comboID)
		}
	}()
}

func GetComboID(keys map[uint32]bool) string {
	var codes []int
	for k := range keys {
		codes = append(codes, int(k))
	}
	// 排序是为了ID唯一
	sort.Ints(codes)

	var idBuilder strings.Builder
	for i, code := range codes {
		if i > 0 {
			idBuilder.WriteString("+")
		}
		idBuilder.WriteString(fmt.Sprintf("%d", code))
	}
	return idBuilder.String()
}

func GetReadableName(keys map[uint32]bool) string {
	uniqueNames := make(map[string]bool)
	var parts []string

	for k := range keys {
		name := getKeyName(k)
		if !uniqueNames[name] {
			uniqueNames[name] = true
			parts = append(parts, name)
		}
	}

	// 排序逻辑：Ctrl, Shift, Alt, Win 在前，其他在后
	sort.Slice(parts, func(i, j int) bool {
		order := map[string]int{
			"Ctrl": 0, "Shift": 1, "Alt": 2, "Win": 3,
		}

		w1, ok1 := order[parts[i]]
		w2, ok2 := order[parts[j]]

		if ok1 && ok2 {
			return w1 < w2
		}
		if ok1 {
			return true
		}
		if ok2 {
			return false
		}
		return parts[i] < parts[j]
	})

	return strings.Join(parts, "+")
}

// 简单的键名映射辅助函数
func getKeyName(vkCode uint32) string {
	switch vkCode {
	// --- 鼠标侧键 ---
	case winapi.VK_XBUTTON1:
		return "MouseBack" // 改个更直观的名字
	case winapi.VK_XBUTTON2:
		return "MouseForward"

	// --- 功能键 ---
	case 0x08:
		return "Back"
	case 0x09:
		return "Tab"
	case 0x0D:
		return "Enter"
	case 0x10, 0xA0, 0xA1:
		return "Shift"
	case 0x11, 0xA2, 0xA3:
		return "Ctrl"
	case 0x12, 0xA4, 0xA5:
		return "Alt"
	case 0x13:
		return "Pause"
	case 0x14:
		return "Caps"
	case 0x1B:
		return "Esc"
	case 0x20:
		return "Space"
	case 0x21:
		return "PgUp"
	case 0x22:
		return "PgDn"
	case 0x23:
		return "End"
	case 0x24:
		return "Home"
	case 0x25:
		return "←"
	case 0x26:
		return "↑"
	case 0x27:
		return "→"
	case 0x28:
		return "↓"
	case 0x2C:
		return "PrtSc"
	case 0x2D:
		return "Ins"
	case 0x2E:
		return "Del"
	case 0x5B, 0x5C:
		return "Win"
	case 0x5D:
		return "Menu"

	// --- 小键盘运算符 ---
	case 0x6A:
		return "Num*"
	case 0x6B:
		return "Num+"
	case 0x6C:
		return "NumEnter" // 有些键盘没有
	case 0x6D:
		return "Num-"
	case 0x6E:
		return "Num."
	case 0x6F:
		return "Num/"

	// --- 主键盘标点符号 (OEM Keys) ---
	// 注意：这些键名基于美式标准键盘
	case 0xBA:
		return ";"
	case 0xBB:
		return "="
	case 0xBC:
		return ","
	case 0xBD:
		return "-"
	case 0xBE:
		return "."
	case 0xBF:
		return "/"
	case 0xC0:
		return "`" // 波浪号键
	case 0xDB:
		return "["
	case 0xDC:
		return "\\"
	case 0xDD:
		return "]"
	case 0xDE:
		return "'"

	default:
		// --- 字母 A-Z ---
		if vkCode >= 'A' && vkCode <= 'Z' {
			return string(rune(vkCode))
		}
		// --- 主键盘数字 0-9 ---
		if vkCode >= '0' && vkCode <= '9' {
			return string(rune(vkCode))
		}
		// --- 小键盘数字 0-9 (VK_NUMPAD0 - VK_NUMPAD9) ---
		if vkCode >= 0x60 && vkCode <= 0x69 {
			return fmt.Sprintf("Num%d", vkCode-0x60)
		}
		// --- F1 - F12 ---
		if vkCode >= 0x70 && vkCode <= 0x7B {
			return fmt.Sprintf("F%d", vkCode-0x6F)
		}
		// --- F13 - F24 (极少用到) ---
		if vkCode >= 0x7C && vkCode <= 0x87 {
			return fmt.Sprintf("F%d", vkCode-0x7C+13)
		}

		// 未知键
		return fmt.Sprintf("Key%d", vkCode)
	}
}
