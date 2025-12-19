package config

import (
	"Q-Solver/pkg/shortcut"
	"encoding/json"
)

type Config struct {
	APIKey             string                         `json:"apiKey"`
	Model              string                         `json:"model"`
	BaseURL            string                         `json:"baseURL"`
	Prompt             string                         `json:"prompt"`
	Opacity            float64                        `json:"opacity"`
	NoCompression      bool                           `json:"noCompression"`
	CompressionQuality int                            `json:"compressionQuality"`
	Sharpening         float64                        `json:"sharpening"`
	Grayscale          bool                           `json:"grayscale"`
	VoiceListening     bool                           `json:"voiceListening"`    // 是否开启语音监听(扬声器)
	KeepContext        bool                           `json:"keepContext"`       // 是否保留上下文
	InterruptThinking  bool                           `json:"interruptThinking"` // 思考时是否打断
	ScreenshotMode     string                         `json:"screenshotMode"`    // "fullscreen" or "window"
	ResumePath         string                         `json:"resumePath"`        // 简历路径
	ResumeBase64       string                         `json:"-"`                 // 简历的 Base64 编码
	ResumeContent      string                         `json:"resumeContent"`     // 简历解析后的 Markdown 内容
	UseMarkdownResume  bool                           `json:"useMarkdownResume"` // 是否使用 Markdown 简历
	Shortcuts          map[string]shortcut.KeyBinding `json:"shortcuts"`
	UseSpeechToText    bool                           `json:"-"` // 是否使用语音转文字,false的话就直接多模态输入
}

// DefaultModel 默认模型
const DefaultModel = "gemini-2.5-flash"

func NewDefaultConfig() Config {
	// 默认配置 - 敏感信息从环境变量读取
	return Config{
		APIKey:            "", // 从环境变量 GHOST_API_KEY 读取，或由用户配置
		Model:             DefaultModel,
		BaseURL:           "",
		ResumePath:        "",
		Prompt:            "",
		Opacity:           1.0,
		VoiceListening:    false,
		KeepContext:       false,
		InterruptThinking: false,    // 默认关闭打断
		ScreenshotMode:    "window", // 默认为窗口区域

		Shortcuts: map[string]shortcut.KeyBinding{
			// F8(119) -> 触发解题
			"solve": {ComboID: "119", KeyName: "F8"},

			// F9(120) -> 切换显示/隐藏
			"toggle": {ComboID: "120", KeyName: "F9"},

			// F10(121) -> 切换鼠标穿透
			"clickthrough": {ComboID: "121", KeyName: "F10"},

			// Alt(164) + Up(38) -> 排序后: 38+164
			"move_up": {ComboID: "38+164", KeyName: "Alt+↑"},

			// Alt(164) + Down(40) -> 排序后: 40+164
			"move_down": {ComboID: "40+164", KeyName: "Alt+↓"},

			// Alt(164) + Left(37) -> 排序后: 37+164
			"move_left": {ComboID: "37+164", KeyName: "Alt+←"},

			// Alt(164) + Right(39) -> 排序后: 39+164
			"move_right": {ComboID: "39+164", KeyName: "Alt+→"},

			// Alt(164) + PgUp(33) -> 排序后: 33+164
			"scroll_up": {ComboID: "33+164", KeyName: "Alt+PgUp"},

			// Alt(164) + PgDn(34) -> 排序后: 34+164
			"scroll_down": {ComboID: "34+164", KeyName: "Alt+PgDn"},
		},
		ResumeBase64:    "",
		UseSpeechToText: false, //暂时先不用
	}
}

func (c *Config) ToJSON() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}
