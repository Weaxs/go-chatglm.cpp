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
	"unsafe"
)

type Chatglm struct {
	pipeline unsafe.Pointer
}

func New(model string) (*Chatglm, error) {
	result := C.load_model(C.CString(model))
	if result == nil {
		return nil, fmt.Errorf("failed loading model")
	}

	llm := &Chatglm{pipeline: result}
	return llm, nil
}

func (llm *Chatglm) Chat(history []string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	out := make([]byte, 1)
	success := C.chat(llm.pipeline, (**C.char)(unsafe.Pointer(&history[0])), C.int(len(history)),
		C.int(opt.MaxLength), C.int(opt.MaxContextLength), C.bool(opt.DoSample), C.int(opt.TopK), C.float(opt.TopP),
		C.float(opt.Temperature), C.float(opt.RepetitionPenalty), C.int(opt.NumThreads), C.bool(opt.Stream),
		(*C.char)(unsafe.Pointer(&out[0])))

	if success != 0 {
		return "", fmt.Errorf("failed loading model")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	return res, nil
}

func (llm *Chatglm) Generate(prompt string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	var out []byte
	result := C.generate(llm.pipeline, C.CString(prompt), C.int(opt.MaxLength), C.int(opt.MaxContextLength), C.bool(opt.DoSample),
		C.int(opt.TopK), C.float(opt.TopP), C.float(opt.Temperature), C.float(opt.RepetitionPenalty),
		C.int(opt.NumThreads), C.bool(opt.Stream), (*C.char)(unsafe.Pointer(&out[0])))

	if result != 0 {
		return "", fmt.Errorf("failed loading model")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	return res, nil

}
