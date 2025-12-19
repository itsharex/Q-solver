package audio

import (
	"Q-Solver/pkg/logger"
	"bytes"
	"context"
	"encoding/binary"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

const (
	// TargetSampleRate 目标采样率 (16kHz 是语音识别的标准)
	TargetSampleRate = 16000
	// TargetChannels 目标声道数 (单声道)
	TargetChannels = 1
	// TargetBits 目标位深 (16位)
	TargetBits = 16

	RingBufferTime  = 1500 * time.Millisecond // 1.5s 预录
	SilenceWindow   = 1000 * time.Millisecond // 静音1s后停止
	EnergyThreshold = 500.0                   // RMS 阈值
	Threshold       = 0.80                    // VAD 判定阈值
)

// AudioRecorder 负责系统音频 Loopback 录制 + 压缩 + VAD
type AudioRecorder struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	isRunning   bool
	isRecording bool

	ringBuffer   *RingBuffer
	recordedData []byte
	silenceStart time.Time
	audioChan    chan *[]byte
	mutex        sync.Mutex

	// Malgo Context
	ctx    *malgo.AllocatedContext
	device *malgo.Device

	OnTranscriptionHandler func(text string, isFinal bool)
	SherpaManager          *SherpaManager
	vadAccumulator         []byte
	startSpeaking          bool //用来判断是否是刚开始说话,用来触发start回调
	EventChan              chan<- AudioEvent
}

var audioBufferPool = sync.Pool{
	New: func() interface{} {
		// 预分配 4KB (足够容纳约 128ms 的 16k/16bit/Mono 音频)
		// 即使不够，后续逻辑也会自动扩容，不用担心
		b := make([]byte, 0, 4096)
		return &b
	},
}

// NewAudioRecorder 创建实例
func NewAudioRecorder(eventChan chan<- AudioEvent) (*AudioRecorder, error) {

	bufferSize := int(TargetSampleRate * TargetChannels * (TargetBits / 8) * (RingBufferTime.Milliseconds() / 1000))
	sherpa, err := NewSherpaManager(TargetSampleRate)
	if err != nil {
		logger.Printf("初始化语音识别失败: %v", err)
		return nil, err
	}

	// 初始化 Malgo 上下文
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		// logger.Printf("MALGO: %v", message)
	})
	if err != nil {
		return nil, err
	}
	return &AudioRecorder{
		ringBuffer:    NewRingBuffer(bufferSize),
		audioChan:     make(chan *[]byte, 1000),
		SherpaManager: sherpa,
		ctx:           ctx,
		startSpeaking: false,
		EventChan:     eventChan,
	}, nil
}

// Start 开始捕获
func (ar *AudioRecorder) Start() error {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	if ar.isRunning {
		logger.Println("音频录制器已在运行中")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	ar.cancel = cancel
	ar.isRunning = true
	ar.isRecording = false
	ar.recordedData = make([]byte, 0)
	ar.ringBuffer.Reset()
	ar.silenceStart = time.Time{}

	// 配置 Loopback 设备
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = TargetChannels
	deviceConfig.SampleRate = TargetSampleRate
	deviceConfig.Alsa.NoMMap = 1

	// 数据回调
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		if framecount == 0 {
			return
		}
		size := len(pInputSamples)

		// 从 Pool 获取 Buffer
		outPtr := audioBufferPool.Get().(*[]byte)
		//不够的话就扩容
		if cap(*outPtr) < size {
			*outPtr = make([]byte, size)
			logger.Printf("扩容音频缓冲区 audioBufferPool 到 %d 字节", size)
		}
		*outPtr = (*outPtr)[:size]

		copy(*outPtr, pInputSamples)

		// 非阻塞发送
		select {
		case ar.audioChan <- outPtr:
		default:
			// 管道满了！这是严重的性能警报，说明消费端卡死了
			logger.Println("音频channel溢出")
			audioBufferPool.Put(outPtr)
		}
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ar.ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		ar.isRunning = false
		cancel()
		return err
	}
	ar.device = device

	if err := device.Start(); err != nil {
		device.Uninit()
		ar.isRunning = false
		cancel()
		return err
	}

	ar.wg.Add(1)
	go ar.processLoop(ctx)
	logger.Println("音频录制器：环回录制已启动（Malgo 16k 单声道）")
	return nil
}

// Stop 停止捕获
// Stop 停止捕获
func (ar *AudioRecorder) Stop() {
	ar.mutex.Lock()
	// 注意：这里不要用 defer Unlock，因为我们要手动控制解锁时机

	if !ar.isRunning {
		ar.mutex.Unlock()
		return
	}

	// 1. 先修改状态，防止新的 Start 进来（或者让新的 Start 知道我们正在停止）
	ar.isRunning = false
	ar.isRecording = false

	// 2. 将需要清理的资源提取到局部变量
	// 这样做是为了让我们能立刻释放锁，同时防止资源在锁外被并发修改
	deviceToStop := ar.device
	cancelFunc := ar.cancel

	// 清空结构体中的引用
	ar.device = nil
	ar.cancel = nil

	// 3. ⚠️ 关键：在执行耗时操作和等待之前，先释放锁！
	ar.mutex.Unlock()

	// --- 以下操作不再持有锁，避免死锁 ---

	// 4. 停止底层设备 (可能耗时)
	if deviceToStop != nil {
		deviceToStop.Uninit()
	}

	// 5. 通知协程退出
	if cancelFunc != nil {
		cancelFunc()
	}

	if ar.SherpaManager != nil {
		ar.SherpaManager.Reset()
	}
	// 6. 等待协程真正结束
	// 此时即使 processLoop 内部请求锁，也不会死锁，因为我们已经 Unlock 了
	ar.wg.Wait()
}

// Close 释放
func (ar *AudioRecorder) Close() {
	ar.Stop()
	if ar.SherpaManager != nil {
		ar.SherpaManager.Close()
	}
	if ar.ctx != nil {
		ar.ctx.Uninit()
		ar.ctx.Free()
		ar.ctx = nil
	}
}

func (ar *AudioRecorder) processLoop(ctx context.Context) {
	defer ar.wg.Done()

	for {
		select {
		case <-ctx.Done():
			logger.Println("音频处理协程已退出")
			return
		case samples := <-ar.audioChan:
			ar.handleAudioSample(*samples)
			audioBufferPool.Put(samples)
		}
	}
}

func (ar *AudioRecorder) handleAudioSample(samples []byte) {

	ar.vadAccumulator = append(ar.vadAccumulator, samples...)

	// 假设我们每 64ms (约2048字节) 或 100ms 处理一次
	// 这里的 bufferSize 只要不是太小(比如几字节)导致频繁调用就行
	const processChunkSize = 1024 * 2

	if len(ar.vadAccumulator) < processChunkSize {
		return
	}

	// 取出数据
	chunkToProcess := ar.vadAccumulator
	// 清空缓冲区 (或者你可以做切片处理，这里为了简化直接全取)
	ar.vadAccumulator = make([]byte, 0, processChunkSize)

	// 直接丢给SherpaManager就好了，不需要vad了，之前太复杂了
	text, isEndpoint, err := ar.SherpaManager.Recognizer(chunkToProcess, TargetSampleRate)
	if err != nil {
		logger.Println("Sherpa 出错:", err)
		return
	}

	//通过channel来进行同步
	if text != "" {
		if !ar.startSpeaking {
			logger.Println("检测到开始说话")
			ar.startSpeaking = true
			// go ar.OnSpeechStart()
			ar.EventChan <- AudioEvent{
				Type:    EventStarted,
				Payload: nil,
			}
		}

		ar.EventChan <- AudioEvent{
			Type:    EventTranscription,
			Payload: TranscriptionData{Text: text, IsFinal: isEndpoint},
		}

		if isEndpoint {
			logger.Println("检测到句尾，识别结果:", text)
			ar.startSpeaking = false
			ar.EventChan <- AudioEvent{
				Type:    EventStopped,
				Payload: chunkToProcess,
			}
		}
	}
}

// encodeWAV 生成标准 16k Mono 16bit WAV 头
func EncodeWAV(pcmData []byte) []byte {
	header := new(bytes.Buffer)

	header.WriteString("RIFF")
	binary.Write(header, binary.LittleEndian, uint32(36+len(pcmData)))
	header.WriteString("WAVE")
	header.WriteString("fmt ")
	binary.Write(header, binary.LittleEndian, uint32(16))               // Subchunk1Size
	binary.Write(header, binary.LittleEndian, uint16(1))                // PCM
	binary.Write(header, binary.LittleEndian, uint16(TargetChannels))   // 1 Channel
	binary.Write(header, binary.LittleEndian, uint32(TargetSampleRate)) // 16000 Hz

	// ByteRate = SampleRate * NumChannels * BitsPerSample/8
	byteRate := uint32(TargetSampleRate * TargetChannels * (TargetBits / 8))
	binary.Write(header, binary.LittleEndian, byteRate)

	// BlockAlign = NumChannels * BitsPerSample/8
	blockAlign := uint16(TargetChannels * (TargetBits / 8))
	binary.Write(header, binary.LittleEndian, blockAlign)

	binary.Write(header, binary.LittleEndian, uint16(TargetBits)) // 16 bits

	header.WriteString("data")
	binary.Write(header, binary.LittleEndian, uint32(len(pcmData)))

	return append(header.Bytes(), pcmData...)
}
