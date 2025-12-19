package audio

// 1. 定义专属类型：避免直接使用 string，提高代码可读性和安全性
type EventType string

// 2. 使用常量块定义事件：方便管理和修改
const (
	EventStarted       EventType = "StartedRecording"
	EventStopped       EventType = "StoppedRecording"
	EventTranscription EventType = "Transcription"
	EventError         EventType = "Error" // 建议增加错误事件
)


//转录的时候发送的数据
type TranscriptionData struct {
	Text    string
	IsFinal bool
}
// 3. 结构体优化
type AudioEvent struct {
	Type    EventType
	Payload any
}

func NewEvent(t EventType, data any) *AudioEvent {
	return &AudioEvent{
		Type:    t,
		Payload: data,
	}
}
