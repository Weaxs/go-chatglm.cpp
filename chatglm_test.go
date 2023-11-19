package chatglm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerate(t *testing.T) {
	chatglm, err := New("./chatglm3-ggml-q4_0.bin")
	if err != nil {
		assert.Fail(t, "load model failed.")
	}

	ret, err := chatglm.Generate("2+2等于多少")
	if err != nil {
		return
	}
	assert.Contains(t, ret, "4")
}

func TestChat(t *testing.T) {
	chatglm, err := New("./chatglm3-ggml-q4_0.bin")
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
