package chatglm

// #cgo CXXFLAGS: -std=c++17
// #cgo CXXFLAGS: -I${SRCDIR}/chatglm.cpp
// #cgo CXXFLAGS: -I${SRCDIR}/chatglm.cpp/third_party/ggml/include/ggml -I${SRCDIR}/chatglm.cpp/third_party/ggml/src
// #cgo CXXFLAGS: -I${SRCDIR}/chatglm.cpp/third_party/sentencepiece/src
// #cgo LDFLAGS: -L${SRCDIR}/ -lbinding -lm -lstdc++
// #cgo darwin LDFLAGS: -framework Accelerate
// #include "binding.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

type Chatglm struct {
	pipeline unsafe.Pointer
	// default stream, of course you can customize stream by  StreamCallback
	stream strings.Builder
}

// New create llm struct
func New(model string) (*Chatglm, error) {
	modelPath := C.CString(model)
	defer C.free(unsafe.Pointer(modelPath))
	result := C.load_model(modelPath)
	if result == nil {
		return nil, fmt.Errorf("failed loading model")
	}

	llm := &Chatglm{pipeline: result}
	return llm, nil
}

// Chat sync chat
func (llm *Chatglm) Chat(history []string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	reverseCount := len(history)
	reversePrompt := make([]*C.char, reverseCount)
	var pass **C.char
	for i, s := range history {
		cs := C.CString(s)
		reversePrompt[i] = cs
		pass = &reversePrompt[0]
	}

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	success := C.chat(llm.pipeline, pass, C.int(reverseCount), params, (*C.char)(unsafe.Pointer(&out[0])))

	if success != 0 {
		return "", fmt.Errorf("model chat failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	res = strings.TrimPrefix(res, " ")
	res = strings.TrimPrefix(res, "\n")
	return res, nil
}

func (llm *Chatglm) StreamChat(history []string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	reverseCount := len(history)
	reversePrompt := make([]*C.char, reverseCount)
	var pass **C.char
	for i, s := range history {
		cs := C.CString(s)
		reversePrompt[i] = cs
		pass = &reversePrompt[0]
	}

	if opt.StreamCallback != nil {
		addStreamCallback(llm.pipeline, opt.StreamCallback)
	} else {
		addStreamCallback(llm.pipeline, defaultStreamCallback(llm))
	}

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	success := C.stream_chat(llm.pipeline, pass, C.int(reverseCount), params, (*C.char)(unsafe.Pointer(&out[0])))
	if success != 0 {
		return "", fmt.Errorf("model chat failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	res = strings.TrimPrefix(res, " ")
	res = strings.TrimPrefix(res, "\n")
	return res, nil
}

func (llm *Chatglm) Generate(prompt string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	result := C.generate(llm.pipeline, C.CString(prompt), params, (*C.char)(unsafe.Pointer(&out[0])))

	if result != 0 {
		return "", fmt.Errorf("model generate failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	return res, nil
}

func (llm *Chatglm) StreamGenerate(prompt string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	if opt.StreamCallback != nil {
		addStreamCallback(llm.pipeline, opt.StreamCallback)
	} else {
		addStreamCallback(llm.pipeline, defaultStreamCallback(llm))
	}

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	result := C.generate(llm.pipeline, C.CString(prompt), params, (*C.char)(unsafe.Pointer(&out[0])))

	if result != 0 {
		return "", fmt.Errorf("model generate failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	return res, nil
}

func (llm *Chatglm) Embeddings(text string, opts ...GenerationOption) ([]int, error) {
	opt := NewGenerationOptions(opts...)
	input := C.CString(text)
	if opt.MaxLength == 0 {
		opt.MaxLength = 99999999
	}
	ints := make([]int, opt.MaxLength)

	params := allocateParams(opt)
	ret := C.get_embedding(llm.pipeline, params, input, (*C.int)(unsafe.Pointer(&ints[0])))
	if ret != 0 {
		return ints, fmt.Errorf("embedding failed")
	}

	return ints, nil
}

func (llm *Chatglm) Free() {
	C.chatglm_free_model(llm.pipeline)
}

func allocateParams(opt *GenerationOptions) unsafe.Pointer {
	return C.chatglm_allocate_params(C.int(opt.MaxLength), C.int(opt.MaxContextLength), C.bool(opt.DoSample),
		C.int(opt.TopK), C.float(opt.TopP), C.float(opt.Temperature), C.float(opt.RepetitionPenalty),
		C.int(opt.NumThreads))
}

func freeParams(params unsafe.Pointer) {
	C.chatglm_free_params(params)
}

var (
	m         sync.RWMutex
	callbacks = map[unsafe.Pointer]func(string) bool{}
)

//export streamCallback
func streamCallback(pipeline unsafe.Pointer, printableText *C.char) C.bool {
	m.RLock()
	defer m.RUnlock()

	if callback, ok := callbacks[pipeline]; ok {
		return C.bool(callback(C.GoString(printableText)))
	}

	return C.bool(true)
}

func addStreamCallback(pipeline unsafe.Pointer, callback func(string) bool) {
	m.Lock()
	defer m.Unlock()

	if callback == nil {
		delete(callbacks, pipeline)
	} else {
		callbacks[pipeline] = callback
	}
}

// return default stream callback
func defaultStreamCallback(llm *Chatglm) func(string) bool {
	return func(text string) bool {
		_, err := llm.stream.WriteString(text)
		if err != nil {
			return false
		}
		return true
	}
}
