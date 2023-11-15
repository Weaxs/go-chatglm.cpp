# go-chatglm.cpp

[chatglm.cpp](https://github.com/li-plus/chatglm.cpp) golang bindings.

The go-chatglm.cpp bindings are high level, as such most of the work is kept into the C/C++ code to avoid any extra computational cost, be more performant and lastly ease out maintenance, while keeping the usage as simple as possible.

# Attention

### Environment

You need to make sure there are `make`, `cmake`, `gcc` command in your machine, otherwise should support C++17.

If you want to run on **Windows OS**, you can use [cygwin](https://www.cygwin.com/).

### Not Support LoRA model
go-chatglm.cpp is not anymore compatible with `LoRA model`, but it woks ONLY with the model which merged by LoRA model and base model.

You can use [convert.py](https://github.com/li-plus/chatglm.cpp/blob/main/chatglm_cpp/convert.py) in [chatglm.cpp](https://github.com/li-plus/chatglm.cpp) 
to merge LoRA model into base model.

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
LIBRARY_PATH=$PWD C_INCLUDE_PATH=$PWD go run ./examples -m "/model/path/here" -t 14
```


# Acknowledgements
 *  This project is greatly inspired by [@mudler](https://github.com/mudler)'s [go-llama.cpp](https://github.com/go-skynet/go-llama.cpp)