package main

import (
	"fmt"
	"github.com/Weaxs/go-chatglm.cpp"
)

func main() {
	llm, err := chatglm.New("./chatglm3-ggml-q4_0.bin")
	if err != nil {
		return
	}

	var history []string
	history = append(history, "你好，我叫 Weaxs")
	res, err := llm.Generate(history[0])
	if err != nil {
		return
	}
	fmt.Println(res)
	history = append(history, res)
	history = append(history, "我的名字是什么")
	res, err = llm.Chat(history)
	if err != nil {
		return
	}
	fmt.Println(res)
}
