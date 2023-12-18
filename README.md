# go-chatglm.cpp

[![GoDoc](https://godoc.org/github.com/Weaxs/go-chatglm.cpp?status.svg)](https://godoc.org/github.com/Weaxs/go-chatglm.cpp)
[![Go Report Card](https://goreportcard.com/badge/github.com/Weaxs/go-chatglm.cpp)](https://goreportcard.com/report/github.com/Weaxs/go-chatglm.cpp)
[![License](https://img.shields.io/github/license/Weaxs/go-chatglm.cpp)](https://github.com/Weaxs/go-chatglm.cpp/blob/main/LICENSE)

[chatglm.cpp](https://github.com/li-plus/chatglm.cpp) golang bindings.

The go-chatglm.cpp bindings are high level, as such most of the work is kept into the C/C++ code to avoid any extra computational cost, be more performant and lastly ease out maintenance, while keeping the usage as simple as possible.

# Attention！

### Environment

You need to make sure there are `make`, `cmake`, `gcc` command in your machine, otherwise should support C++17.

If you want to run on **Windows OS**, you can use [cygwin](https://www.cygwin.com/) or [MinGW](https://www.mingw-w64.org/).

> **`cmake` > 3.8**  and  **`gcc` > 5.1.0**  (support C++17)

### Not Support LoRA model

go-chatglm.cpp is not anymore compatible with `LoRA model`, but it woks ONLY with the model which merged by LoRA model and base model.

You can use [convert.py](https://github.com/li-plus/chatglm.cpp/blob/main/chatglm_cpp/convert.py) in [chatglm.cpp](https://github.com/li-plus/chatglm.cpp) to merge LoRA model into base model.

# Usage

Note: This repository uses git submodules to keep track of [chatglm.cpp](https://github.com/li-plus/chatglm.cpp) .

Clone the repository locally:

```shell
git clone --recurse-submodules https://github.com/Weaxs/go-chatglm.cpp
```

To build the bindings locally, run:

```shell
cd go-chatglm.cpp
make libbinding.a
```

Now you can run the example with:

```shell
go run ./examples -m "/model/path/here"
                    ____ _           _    ____ _     __  __                   
  __ _  ___        / ___| |__   __ _| |_ / ___| |   |  \/  |  ___ _ __  _ __  
 / _` |/ _ \ _____| |   | '_ \ / _` | __| |  _| |   | |\/| | / __| '_ \| '_ \ 
| (_| | (_) |_____| |___| | | | (_| | |_| |_| | |___| |  | || (__| |_) | |_) |
 \__, |\___/       \____|_| |_|\__,_|\__|\____|_____|_|  |_(_)___| .__/| .__/ 
 |___/                                                           |_|   |_|    

>>> 你好

Sending 你好


你好👋！我是人工智能助手 ChatGLM3-6B，很高兴见到你，欢迎问我任何问题。
```

# Acceleration

## Metal (Apple Silicon)

MPS (Metal Performance Shaders) allows computation to run on powerful Apple Silicon GPU.

```
BUILD_TYPE=metal make libbinding.a
go build -tags metal ./examples/main.go
./main -m "/model/path/here"
```

## OpenBLAS

OpenBLAS provides acceleration on CPU.

```
BUILD_TYPE=openblas make libbinding.a
go build -tags openblas ./examples/main.go
./main -m "/model/path/here"
```

## cuBLAS

cuBLAS uses NVIDIA GPU to accelerate BLAS.

```
BUILD_TYPE=cublas make libbinding.a
go build -tags cublas ./examples/main.go
./main -m "/model/path/here"
```

# Acknowledgements

* This project is greatly inspired by [@mudler](https://github.com/mudler)'s [go-llama.cpp](https://github.com/go-skynet/go-llama.cpp)
