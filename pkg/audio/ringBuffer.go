package audio

import "sync"

// --- RingBuffer ---
type RingBuffer struct {
	buffer []byte
	size   int
	pos    int
	full   bool
	mutex  sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer: make([]byte, size),
		size:   size,
	}
}

func (rb *RingBuffer) Reset() {
	rb.pos = 0
	rb.full = false
}

func (rb *RingBuffer) Write(data []byte) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	n := len(data)
	//数据超出环形缓冲区，直接截取最后一段填满
	if n > rb.size {
		copy(rb.buffer, data[n-rb.size:])
		rb.pos = 0
		rb.full = true
		return
	}
	//空间足够，直接写入
	if rb.pos+n <= rb.size {
		copy(rb.buffer[rb.pos:], data)
		rb.pos += n
	} else {
		firstChunk := rb.size - rb.pos
		copy(rb.buffer[rb.pos:], data[:firstChunk])
		copy(rb.buffer[0:], data[firstChunk:])
		rb.pos = n - firstChunk
		rb.full = true
	}
	if rb.pos == rb.size {
		rb.pos = 0
		rb.full = true
	}
}

func (rb *RingBuffer) ReadAll() []byte {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	if !rb.full {
		return append([]byte(nil), rb.buffer[:rb.pos]...)
	}
	result := make([]byte, rb.size)
	copy(result, rb.buffer[rb.pos:])
	copy(result[rb.size-rb.pos:], rb.buffer[:rb.pos])
	return result
}

// IsEmpty 判断当前缓冲区是否为空
func (rb *RingBuffer) IsEmpty() bool {
    rb.mutex.Lock()
    defer rb.mutex.Unlock()

    // 只有当 "未满" 且 "写入位置在0" 时，才视为空
    // 如果 full=true 但 pos=0，说明刚写满一圈回到起点，此时不算空
    return !rb.full && rb.pos == 0
}