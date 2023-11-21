package chatglm

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var chatglm *Chatglm

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

func TestChat(t *testing.T) {
	history := []string{"2+2等于多少"}
	ret, err := chatglm.Chat(history)
	if err != nil {
		assert.Fail(t, "first chat failed")
	}
	assert.Contains(t, ret, "4")

	history = append(history, ret)
	history = append(history, "再加4等于多少")
	ret, err = chatglm.Chat(history)
	if err != nil {
		assert.Fail(t, "second chat failed")
	}
	assert.Contains(t, ret, "8")

	history = append(history, ret)
	assert.Len(t, history, 4)
}

func TestEmbedding(t *testing.T) {
	maxLength := 1024
	embeddings, err := chatglm.Embeddings("你好", SetMaxLength(1024))
	if err != nil {
		assert.Fail(t, "embedding failed.")
	}
	assert.Len(t, embeddings, maxLength)
}
