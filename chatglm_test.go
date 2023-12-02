package chatglm

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var (
	chatglm   *Chatglm
	modelType string
)

func setup() {
	testModelPath, exist := os.LookupEnv("TEST_MODEL")
	if !exist {
		testModelPath = "./chatglm3-ggml-q4_0.bin"
	}

	var err error
	chatglm, err = New(testModelPath)
	if err != nil {
		panic("load model failed.")
	}
	modelType = chatglm.ModelType()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	defer chatglm.Free()
	os.Exit(code)
}

func TestGenerate(t *testing.T) {
	ret, err := chatglm.Generate("2+2等于多少")
	if err != nil {
		assert.Fail(t, "generate failed.")
	}
	assert.Contains(t, ret, "4")
}

func TestStreamGenerate(t *testing.T) {
	prompt := "2+2等于多少"
	ret, err := chatglm.StreamGenerate(prompt)
	if err != nil {
		assert.Fail(t, "stream generate failed.")
	}
	streamOut := chatglm.stream.String()
	defer chatglm.stream.Reset()

	assert.Contains(t, streamOut, "4")
	assert.Contains(t, ret, "4")
}

func TestChat(t *testing.T) {
	var messages []*ChatMessage
	messages = append(messages, NewUserMsg("2+2等于多少"))
	ret, err := chatglm.Chat(messages)
	if err != nil {
		assert.Fail(t, "first chat failed")
	}
	assert.Contains(t, ret, "4")

	messages = append(messages, NewAssistantMsg(ret, modelType))
	messages = append(messages, NewUserMsg("再加4等于多少"))
	ret, err = chatglm.Chat(messages)
	if err != nil {
		assert.Fail(t, "second chat failed")
	}
	assert.Contains(t, ret, "8")

	messages = append(messages, NewAssistantMsg(ret, modelType))
	assert.Len(t, messages, 4)
}

func TestChatStream(t *testing.T) {
	var messages []*ChatMessage
	messages = append(messages, NewUserMsg("2+2等于多少"))
	out1 := strings.Builder{}
	ret, err := chatglm.StreamChat(messages, SetStreamCallback(func(s string) bool {
		out1.WriteString(s)
		return true
	}))
	if err != nil {
		assert.Fail(t, "first chat failed")
	}
	outStr1 := out1.String()
	outStr1 = strings.TrimPrefix(outStr1, " ")
	outStr1 = strings.TrimPrefix(outStr1, "\n")
	assert.Contains(t, ret, "4")
	assert.Contains(t, outStr1, "4")

	messages = append(messages, NewAssistantMsg(ret, modelType))
	messages = append(messages, NewUserMsg("再加4等于多少"))
	ret, err = chatglm.StreamChat(messages)
	if err != nil {
		assert.Fail(t, "second chat failed")
	}
	out2 := chatglm.stream.String()
	out2 = strings.TrimPrefix(out2, " ")
	out2 = strings.TrimPrefix(out2, "\n")
	assert.Contains(t, ret, "8")
	assert.Contains(t, out2, "8")

	messages = append(messages, NewAssistantMsg(ret, modelType))
	assert.Len(t, messages, 4)
}

func TestEmbedding(t *testing.T) {
	maxLength := 1024
	embeddings, err := chatglm.Embeddings("你好", SetMaxLength(maxLength))
	if err != nil {
		assert.Fail(t, "embedding failed.")
	}
	assert.Len(t, embeddings, maxLength)
}

func TestSystemToolCall(t *testing.T) {
	file, err := os.ReadFile("examples/system/function_call.txt")
	if err != nil {
		return
	}
	var messages []*ChatMessage
	messages = append(messages, NewSystemMsg(string(file)))
	messages = append(messages, NewUserMsg("生成一个随机数"))

	ret, err := chatglm.Chat(messages, SetDoSample(false))
	if err != nil {
		assert.Fail(t, "call system tool failed.")
	}
	assert.Contains(t, ret, "```python\ntool_call(seed=42, range=(0, 100))\n```")
	messages = append(messages, NewAssistantMsg(ret, modelType))
	messages = append(messages, NewObservationMsg("22"))

	ret, err = chatglm.Chat(messages, SetDoSample(false))
	if err != nil {
		assert.Fail(t, "call system tool failed.")
	}
	assert.Contains(t, ret, "22")
}

func TestCodeInterpreter(t *testing.T) {
	file, err := os.ReadFile("examples/system/code_interpreter.txt")
	if err != nil {
		return
	}
	var messages []*ChatMessage
	messages = append(messages, NewSystemMsg(string(file)))
	messages = append(messages, NewUserMsg("列出100以内的所有质数"))
	ret, err := chatglm.Chat(messages, SetDoSample(false))
	if err != nil {
		assert.Fail(t, "call code interpreter failed.")
	}
	messages = append(messages, NewAssistantMsg(ret, modelType))
	assert.Contains(t, ret, "好的，我会为您列出100以内的所有质数。\\n\\n质数是指只能被1和它本身整除的大于1的整数。例如，2、3、5、7等都是质数。\\n\\n让我们开始吧！")
	messages = append(messages, NewObservationMsg("[2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97]"))

	ret, err = chatglm.Chat(messages, SetDoSample(false))
	if err != nil {
		assert.Fail(t, "call code interpreter failed.")
	}
	assert.Contains(t, ret, "2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97")
}
