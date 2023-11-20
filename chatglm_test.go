package chatglm

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	testModelPath, exist := os.LookupEnv("TEST_MODEL")
	if !exist {
		testModelPath = "./chatglm3-ggml-q4_0.bin"
	}

	chatglm, err := New(testModelPath)
	defer chatglm.Free()
	if err != nil {
		assert.Fail(t, "load model failed.")
	}

	ret, err := chatglm.Generate("2+2等于多少")
	if err != nil {
		assert.Fail(t, "generate failed.")
	}
	assert.Contains(t, ret, "4")
}

func TestStreamGenerate(t *testing.T) {
	testModelPath, exist := os.LookupEnv("TEST_MODEL")
	if !exist {
		testModelPath = "./chatglm3-ggml-q4_0.bin"
	}

	chatglm, err := New(testModelPath)
	defer chatglm.Free()
	if err != nil {
		assert.Fail(t, "load model failed.")
	}

	err = chatglm.StreamGenerate("2+2等于多少")
	if err != nil {
		assert.Fail(t, "stream generate failed.")
	}

	ret, err := chatglm.GetStream()
	if err != nil {
		assert.Fail(t, "get stream failed.")
	}
	assert.Contains(t, ret, "4")
}

func TestChat(t *testing.T) {
	testModelPath, exist := os.LookupEnv("TEST_MODEL")
	if !exist {
		testModelPath = "./chatglm3-ggml-q4_0.bin"
	}

	chatglm, err := New(testModelPath)
	defer chatglm.Free()
	if err != nil {
		assert.Fail(t, "load model failed.")
	}

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

func TestStreamChat(t *testing.T) {
	testModelPath, exist := os.LookupEnv("TEST_MODEL")
	if !exist {
		testModelPath = "./chatglm3-ggml-q4_0.bin"
	}

	chatglm, err := New(testModelPath)
	defer chatglm.Free()
	if err != nil {
		assert.Fail(t, "load model failed.")
	}

	history := []string{"2+2等于多少"}
	err = chatglm.StreamChat(history)
	if err != nil {
		assert.Fail(t, "first chat failed")
	}
	ret, err := chatglm.GetStream()
	if err != nil {
		assert.Fail(t, "first get stream failed.")
	}
	assert.Contains(t, ret, "4")

	history = append(history, ret)
	history = append(history, "再加4等于多少")
	err = chatglm.StreamChat(history)
	if err != nil {
		assert.Fail(t, "second chat failed")
	}
	ret, err = chatglm.GetStream()
	if err != nil {
		assert.Fail(t, "first get stream failed.")
	}
	assert.Contains(t, ret, "8")

	history = append(history, ret)
	assert.Len(t, history, 4)
}

func TestEmbedding(t *testing.T) {
	testModelPath, exist := os.LookupEnv("TEST_MODEL")
	if !exist {
		testModelPath = "./chatglm3-ggml-q4_0.bin"
	}

	chatglm, err := New(testModelPath)
	defer chatglm.Free()
	if err != nil {
		assert.Fail(t, "load model failed.")
	}

	maxLength := 1024
	embeddings, err := chatglm.Embeddings("你好", SetMaxLength(1024))
	if err != nil {
		assert.Fail(t, "embedding failed.")
	}
	assert.Len(t, embeddings, maxLength)
}
