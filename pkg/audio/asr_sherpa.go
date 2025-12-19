package audio

import (
	"Q-Solver/pkg/logger"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// ErrModelNotFound 模型文件不存在错误
var ErrModelNotFound = errors.New("语音识别模型文件不存在")

type SherpaManager struct {
	recognizer *sherpa.OnlineRecognizer
	stream     *sherpa.OnlineStream
}

// checkModelFiles 检查模型文件是否存在
func checkModelFiles(modelDir string) error {
	requiredFiles := []string{
		"encoder.int8.onnx",
		"decoder.onnx",
		"joiner.int8.onnx",
		"tokens.txt",
	}

	// 检查目录是否存在
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		return ErrModelNotFound
	}

	// 检查所有必需文件
	for _, file := range requiredFiles {
		filePath := filepath.Join(modelDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return ErrModelNotFound
		}
	}

	return nil
}

func NewSherpaManager(sampleRate int) (*SherpaManager, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)

	// 拼接 models 目录绝对路径
	modelDir := filepath.Join(exeDir, "models")

	// 检查模型文件是否存在
	if err := checkModelFiles(modelDir); err != nil {
		logger.Printf("语音模型检查失败: %v，模型目录: %s", err, modelDir)
		return nil, err
	}

	config := sherpa.OnlineRecognizerConfig{
		FeatConfig: sherpa.FeatureConfig{
			SampleRate: sampleRate,
			FeatureDim: 80,
		},
		ModelConfig: sherpa.OnlineModelConfig{
			Transducer: sherpa.OnlineTransducerModelConfig{
				Encoder: modelDir + "/encoder.int8.onnx",
				Decoder: modelDir + "/decoder.onnx",
				Joiner:  modelDir + "/joiner.int8.onnx",
			},
			Tokens:     modelDir + "/tokens.txt",
			NumThreads: 4,
			Debug:      0,
			ModelType:  "zipformer2",
			Provider:   "cpu",
		},
		DecodingMethod:          "greedy_search",
		EnableEndpoint:          1,
		Rule1MinTrailingSilence: 2.4,

		// 规则2: 长语音后的等待时间 (单位: 秒)
		Rule2MinTrailingSilence: 1.2,

		// 规则3: 强制切断的最长句长 (单位: 秒)
		Rule3MinUtteranceLength: 120.0,
	}
	//创建一个识别器
	recognizer := sherpa.NewOnlineRecognizer(&config)
	//创建一个流
	stream := sherpa.NewOnlineStream(recognizer)
	logger.Println("Sherpa-ONNX 识别器初始化完成!")
	return &SherpaManager{
		recognizer: recognizer,
		stream:     stream,
	}, nil
}

func (s *SherpaManager) Recognizer(pcmData []byte, sampleRate int) (string, bool, error) {
	s.stream.AcceptWaveform(sampleRate, BytesToFloat32(pcmData))
	for s.recognizer.IsReady(s.stream) {
		s.recognizer.Decode(s.stream)
	}
	result := s.recognizer.GetResult(s.stream)

	if s.recognizer.IsEndpoint(s.stream) {
		s.recognizer.Reset(s.stream)
		return result.Text, true, nil
	}
	return result.Text, false, nil
}

func (s *SherpaManager) Reset() {
	s.recognizer.Reset(s.stream)
}
func (s *SherpaManager) Close() {
	sherpa.DeleteOnlineStream(s.stream)
	sherpa.DeleteOnlineRecognizer(s.recognizer)
	logger.Println("Sherpa-ONNX 识别器已释放内存。")
}

func BytesToFloat32(data []byte) []float32 {
	// 1. 计算采样点数量 (2个byte = 1个采样点)
	numSamples := len(data) / 2
	samples := make([]float32, numSamples)

	// 2. 遍历转换
	// 这里使用 binary.LittleEndian 是最安全且兼容性最好的方式
	for i := 0; i < numSamples; i++ {
		// 取出 2 个 byte 转成 int16
		valInt16 := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))

		// 归一化到 -1.0 ~ 1.0
		samples[i] = float32(valInt16) / 32768.0
	}

	return samples
}
