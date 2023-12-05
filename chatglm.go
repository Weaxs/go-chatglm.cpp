package chatglm

// #cgo CXXFLAGS: -std=c++17
// #cgo CXXFLAGS: -I${SRCDIR}/chatglm.cpp
// #cgo CXXFLAGS: -I${SRCDIR}/chatglm.cpp/third_party/ggml/include/ggml -I${SRCDIR}/chatglm.cpp/third_party/ggml/src
// #cgo CXXFLAGS: -I${SRCDIR}/chatglm.cpp/third_party/sentencepiece/src
// #cgo LDFLAGS: -L${SRCDIR}/ -lbinding -lm -v
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

func NewAssistantMsg(input string, modelType string) *ChatMessage {
	result := &ChatMessage{Role: RoleAssistant, Content: input}
	if modelType != "ChatGLM3" {
		return result
	}

	if !strings.Contains(input, DELIMITER) {
		return result
	}

	ciPos := strings.Index(input, DELIMITER)
	if ciPos != 0 {
		content := input[:ciPos]
		code := input[ciPos+len(DELIMITER):]
		toolCalls := []*ToolCallMessage{{Type: TypeCode, Code: &CodeMessage{code}}}
		result.Content = content
		result.ToolCalls = toolCalls
	}
	return result
}

func NewUserMsg(content string) *ChatMessage {
	return &ChatMessage{Role: RoleUser, Content: content}
}

func NewSystemMsg(content string) *ChatMessage {
	return &ChatMessage{Role: RoleSystem, Content: content}
}

func NewObservationMsg(content string) *ChatMessage {
	return &ChatMessage{Role: RoleObservation, Content: content}
}

// Chat by history [synchronous]
func (llm *Chatglm) Chat(messages []*ChatMessage, opts ...GenerationOption) (string, error) {
	err := checkChatMessages(messages)
	if err != nil {
		return "", err
	}
	reverseMsgs, err := allocateChatMessages(messages)
	if err != nil {
		return "", err
	}
	reverseCount := len(reverseMsgs)
	pass := &reverseMsgs[0]

	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	success := C.chat(llm.pipeline, pass, C.int(reverseCount), params, (*C.char)(unsafe.Pointer(&out[0])))

	if success != 0 {
		return "", fmt.Errorf("model chat failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	res = removeSpecialTokens(res)
	return res, nil
}

// StreamChat chat with stream output by StreamCallback
func (llm *Chatglm) StreamChat(messages []*ChatMessage, opts ...GenerationOption) (string, error) {
	err := checkChatMessages(messages)
	if err != nil {
		return "", err
	}
	reverseMsgs, err := allocateChatMessages(messages)
	if err != nil {
		return "", err
	}
	reverseCount := len(reverseMsgs)
	pass := &reverseMsgs[0]

	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	if opt.StreamCallback != nil {
		setStreamCallback(llm.pipeline, opt.StreamCallback)
	} else {
		setStreamCallback(llm.pipeline, defaultStreamCallback(llm))
	}
	defer setStreamCallback(llm.pipeline, nil)

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	success := C.stream_chat(llm.pipeline, pass, C.int(reverseCount), params, (*C.char)(unsafe.Pointer(&out[0])))
	if success != 0 {
		return "", fmt.Errorf("model chat failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	res = removeSpecialTokens(res)
	return res, nil
}

// Generate by prompt [synchronous]
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
	res = strings.TrimPrefix(res, " ")
	res = strings.TrimPrefix(res, "\n")
	return res, nil
}

// StreamGenerate with stream output by StreamCallback
func (llm *Chatglm) StreamGenerate(prompt string, opts ...GenerationOption) (string, error) {
	opt := NewGenerationOptions(opts...)
	params := allocateParams(opt)
	defer freeParams(params)

	if opt.StreamCallback != nil {
		setStreamCallback(llm.pipeline, opt.StreamCallback)
	} else {
		setStreamCallback(llm.pipeline, defaultStreamCallback(llm))
	}
	defer setStreamCallback(llm.pipeline, nil)

	if opt.MaxContextLength == 0 {
		opt.MaxContextLength = 99999999
	}
	out := make([]byte, opt.MaxContextLength)
	result := C.stream_generate(llm.pipeline, C.CString(prompt), params, (*C.char)(unsafe.Pointer(&out[0])))

	if result != 0 {
		return "", fmt.Errorf("model generate failed")
	}
	res := C.GoString((*C.char)(unsafe.Pointer(&out[0])))
	res = strings.TrimPrefix(res, " ")
	res = strings.TrimPrefix(res, "\n")
	return res, nil
}

// Embeddings get text input_ids,
func (llm *Chatglm) Embeddings(text string, opts ...GenerationOption) ([]int, error) {
	opt := NewGenerationOptions(opts...)
	input := C.CString(text)
	if opt.MaxLength == 0 {
		opt.MaxLength = 99999999
	}
	ints := make([]int, opt.MaxLength)

	ret := C.get_embedding(llm.pipeline, input, C.int(opt.MaxLength), (*C.int)(unsafe.Pointer(&ints[0])))
	if ret != 0 {
		return ints, fmt.Errorf("embedding failed")
	}

	return ints, nil
}

func (llm *Chatglm) Free() {
	C.free_model(llm.pipeline)
}

func (llm *Chatglm) ModelType() string {
	return C.GoString(C.get_model_type(llm.pipeline))
}

// allocateParams create GenerationOptions from c
func allocateParams(opt *GenerationOptions) unsafe.Pointer {
	return C.allocate_params(C.int(opt.MaxLength), C.int(opt.MaxContextLength), C.bool(opt.DoSample),
		C.int(opt.TopK), C.float(opt.TopP), C.float(opt.Temperature), C.float(opt.RepetitionPenalty),
		C.int(opt.NumThreads))
}

// freeParams
func freeParams(params unsafe.Pointer) {
	C.free_params(params)
}

// checkChatMessages check messages format
func checkChatMessages(messages []*ChatMessage) error {
	n := len(messages)
	if n < 1 {
		return fmt.Errorf("invalid chat messages size: %d", n)
	}
	isSys := messages[0].Role == RoleSystem

	if !isSys && n%2 == 0 {
		return fmt.Errorf("invalid chat messages size: %d", n)
	}
	if isSys && n%2 == 1 {
		return fmt.Errorf("invalid chat messages size: %d", n)
	}

	for i, message := range messages {
		if message.ToolCalls == nil {
			continue
		}

		for j, toolCall := range message.ToolCalls {
			if toolCall.Type == TypeCode && toolCall.Code == nil {
				return fmt.Errorf("expect messages[%d].ToolCalls[%d].Code is not nil", i, j)
			}
			if toolCall.Type == TypeFunction && toolCall.Function == nil {
				return fmt.Errorf("expect messages[%d].ToolCalls[%d].Function is not nil", i, j)
			}
		}
	}
	return nil
}

// allocateChatMessages covert []*ChatMessage in go to []C.ChatMessage in c++
func allocateChatMessages(messages []*ChatMessage) ([]unsafe.Pointer, error) {
	reverseMessages := make([]unsafe.Pointer, len(messages))
	for i, message := range messages {
		var reverseToolCalls []unsafe.Pointer
		if message.ToolCalls != nil {
			for _, toolCall := range message.ToolCalls {
				var codeOrFunc unsafe.Pointer
				if toolCall.Type == TypeCode {
					codeOrFunc = C.create_code(C.CString(toolCall.Code.Input))
				} else if toolCall.Type == TypeFunction {
					codeOrFunc = C.create_function(
						C.CString(toolCall.Function.Name), C.CString(toolCall.Function.Arguments))
				}
				toolCallPoint := C.create_tool_call(C.CString(toolCall.Type), codeOrFunc)
				if toolCallPoint != nil {
					reverseToolCalls = append(reverseToolCalls, toolCallPoint)
				}
			}
		}
		var pass *unsafe.Pointer
		if len(reverseToolCalls) > 0 {
			pass = &reverseToolCalls[0]
		}
		reverseMessages[i] = C.create_chat_message(
			C.CString(message.Role), C.CString(message.Content), pass, C.int(len(reverseToolCalls)))
	}
	return reverseMessages, nil
}

func removeSpecialTokens(data string) string {
	output := strings.ReplaceAll(data, "[MASK]", "")
	output = strings.ReplaceAll(output, "[gMASK]", "")
	output = strings.ReplaceAll(output, "[sMASK]", "")
	output = strings.ReplaceAll(output, "sop", "")
	output = strings.ReplaceAll(output, "eop", "")
	output = strings.Replace(output, "<|assistant|>", "", 1)
	output = strings.TrimSuffix(output, "<|assistant|>")
	output = strings.ReplaceAll(output, "<|assistant|>", DELIMITER)
	output = strings.TrimLeftFunc(output, func(r rune) bool {
		return r == '\n' || r == ' '
	})
	return output
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

// setStreamCallback add callback into global map callbacks
func setStreamCallback(pipeline unsafe.Pointer, callback func(string) bool) {
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
