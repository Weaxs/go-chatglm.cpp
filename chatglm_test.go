package chatglm

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
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

func TestChatStream(t *testing.T) {
	history := []string{"2+2等于多少"}
	out1 := strings.Builder{}
	ret, err := chatglm.StreamChat(history, SetStreamCallback(func(s string) bool {
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

	history = append(history, ret)
	history = append(history, "再加4等于多少")
	ret, err = chatglm.StreamChat(history)
	if err != nil {
		assert.Fail(t, "second chat failed")
	}
	out2 := chatglm.stream.String()
	out2 = strings.TrimPrefix(out2, " ")
	out2 = strings.TrimPrefix(out2, "\n")
	assert.Contains(t, ret, "8")
	assert.Contains(t, out2, "8")

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
