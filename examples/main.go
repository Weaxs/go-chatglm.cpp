package main

import (
	"bufio"
	"flag"
	"fmt"
	c "github.com/Weaxs/go-chatglm.cpp"
	"io"
	"os"
	"strings"
)

func main() {
	var model string
	var system string
	var temp float64
	var topK int
	var topP float64
	var maxLength int
	var maxContentLength int
	var threads int
	var repeatPenalty float64

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&model, "m", "./chatglm3-ggml-q4_0.bin", "path to model file to load")
	flags.StringVar(&system, "s", "", "system message to set the behavior of the assistant")
	flags.Float64Var(&temp, "temp", 0.95, "temperature (default: 0.95)")
	flags.IntVar(&maxLength, "max_length ", 2048, "max total length including prompt and output (default: 2048)")
	flags.IntVar(&maxContentLength, "max_context_length ", 512, " max context length (default: 512)")
	flags.IntVar(&topK, "top_k", 0, "top-k sampling (default: 0)")
	flags.Float64Var(&topP, "top_p", 0.7, "top-p sampling (default: 0.7)")
	flags.Float64Var(&repeatPenalty, "repeat_penalty", 1.0, "penalize repeat sequence of tokens (default: 1.0, 1.0 = disabled)")
	flags.IntVar(&threads, "threads", 0, " number of threads for inference")

	err := flags.Parse(os.Args[1:])

	chatglm, err := c.New(model)
	modelType := chatglm.ModelType()

	if err != nil {
		fmt.Printf("Parsing program arguments failed: %s", err)
		os.Exit(1)
	}

	fmt.Printf("                    ____ _           _    ____ _     __  __                   \n" +
		"  __ _  ___        / ___| |__   __ _| |_ / ___| |   |  \\/  |  ___ _ __  _ __  \n" +
		" / _` |/ _ \\ _____| |   | '_ \\ / _` | __| |  _| |   | |\\/| | / __| '_ \\| '_ \\ \n" +
		"| (_| | (_) |_____| |___| | | | (_| | |_| |_| | |___| |  | || (__| |_) | |_) |\n" +
		" \\__, |\\___/       \\____|_| |_|\\__,_|\\__|\\____|_____|_|  |_(_)___| .__/| .__/ \n" +
		" |___/                                                           |_|   |_|    \n\n")

	reader := bufio.NewReader(os.Stdin)
	var message []*c.ChatMessage

	for {
		text := readMultiLineInput(reader)
		message = append(message, c.NewUserMsg(text))
		r, err := chatglm.StreamChat(message,
			c.SetTemperature(float32(temp)), c.SetTopP(float32(topP)), c.SetTopK(topK),
			c.SetMaxLength(maxLength), c.SetMaxContextLength(maxContentLength),
			c.SetRepetitionPenalty(float32(repeatPenalty)), c.SetNumThreads(threads),
			c.SetStreamCallback(callback))
		if err != nil {
			panic(err)
		}
		message = append(message, c.NewAssistantMsg(r, modelType))

		_, err = chatglm.Embeddings(text)
		if err != nil {
			fmt.Printf("Embeddings: error %s \n", err.Error())
		}
		fmt.Printf("\n\n")
	}

}

func callback(s string) bool {
	fmt.Print(s)
	return true
}

// readMultiLineInput reads input until an empty line is entered.
func readMultiLineInput(reader *bufio.Reader) string {
	var lines []string
	fmt.Print(">>> ")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			fmt.Printf("Reading the prompt failed: %s", err)
			os.Exit(1)
		}

		if len(strings.TrimSpace(line)) == 0 {
			break
		}

		lines = append(lines, line)
	}

	text := strings.Join(lines, "")
	fmt.Println("Sending", text)
	return text
}
