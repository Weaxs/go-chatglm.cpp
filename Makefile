INCLUDE_PATH := $(abspath ./)
LIBRARY_PATH := $(abspath ./)

ifndef UNAME_S
	ifeq ($(OS),Windows_NT)
		UNAME_S := $(shell ver)
	else
		UNAME_S := $(shell uname -s)
	endif
endif

ifndef UNAME_P
	ifeq ($(OS),Windows_NT)
		UNAME_P := $(shell wmic cpu get caption)
	else
		UNAME_P := $(shell uname -p)
	endif
endif

ifndef UNAME_M
	ifeq ($(OS),Windows_NT)
		UNAME_M := $(PROCESSOR_ARCHITECTURE)
	else
		UNAME_M := $(shell uname -m)
	endif
endif


ifeq ($(OS),Windows_NT)
	CP := xcopy
	DELIMITER := \\
else
	CP := cp
	DELIMITER := /
endif

CCV := $(shell $(CC) --version | head -n 1)
CXXV := $(shell $(CXX) --version | head -n 1)

# Mac OS + Arm can report x86_64
# ref: https://github.com/ggerganov/whisper.cpp/issues/66#issuecomment-1282546789
ifeq ($(UNAME_S),Darwin)
	ifneq ($(UNAME_P),arm)
		SYSCTL_M := $(shell sysctl -n hw.optional.arm64 2>/dev/null)
		ifeq ($(SYSCTL_M),1)
			UNAME_P := arm
			UNAME_M := arm64
			warn := $(warning Your arch is announced as x86_64, but it seems to actually be ARM64. Not fixing that can lead to bad performance. For more info see: https://github.com/ggerganov/whisper.cpp/issues/66\#issuecomment-1282546789)
		endif
	endif
endif

#
# Compile flags
#

BUILD_TYPE?=
# keep standard at C17 and C++17
CXXFLAGS = -I. -O3 -DNDEBUG -std=c++17 -fPIC -pthread
CMAKE_ARGS=

# warnings
ifneq ($(OS),Windows_NT)
	CXXFLAGS += -g -Wall -Wextra -Wpedantic -Wcast-qual -Wno-unused-function -pedantic-errors
endif

# GPGPU specific
GGML_CUDA_OBJ_PATH=third_party/ggml/src/CMakeFiles/ggml.dir/ggml-cuda.cu.o


# Architecture specific
# feel free to update the Makefile for your architecture and send a pull request or issue
ifeq ($(UNAME_M),$(filter $(UNAME_M),x86_64 i686))
	# Use all CPU extensions that are available:
	CXXFLAGS += -march=native -mtune=native
endif
ifneq ($(filter ppc64%,$(UNAME_M)),)
	POWER9_M := $(shell grep "POWER9" /proc/cpuinfo)
	ifneq (,$(findstring POWER9,$(POWER9_M)))
		CXXFLAGS += -mcpu=power9
	endif
	# Require c++23's std::byteswap for big-endian support.
	ifeq ($(UNAME_M),ppc64)
		CXXFLAGS += -std=c++23 -DGGML_BIG_ENDIAN
	endif
endif
ifdef CHATGLM_GPROF
	CXXFLAGS += -pg
endif
ifneq ($(filter aarch64%,$(UNAME_M)),)
	CXXFLAGS += -mcpu=native
endif
ifneq ($(filter armv6%,$(UNAME_M)),)
	# Raspberry Pi 1, 2, 3
	CXXFLAGS += -mfpu=neon-fp-armv8 -mfp16-format=ieee -mno-unaligned-access
endif
ifneq ($(filter armv7%,$(UNAME_M)),)
	# Raspberry Pi 4
	CXXFLAGS += -mfpu=neon-fp-armv8 -mfp16-format=ieee -mno-unaligned-access -funsafe-math-optimizations
endif
ifneq ($(filter armv8%,$(UNAME_M)),)
	# Raspberry Pi 4
	CXXFLAGS += -mfp16-format=ieee -mno-unaligned-access
endif

ifeq ($(BUILD_TYPE),cublas)
	CMAKE_ARGS+=-DGGML_CUBLAS=ON
endif
ifeq ($(BUILD_TYPE),openblas)
	CMAKE_ARGS+=-DGGML_OPENBLAS=ON
	CXXFLAGS  += -I/usr/local/include/openblas -lopenblas
    CGO_TAGS=-tags openblas
endif
ifeq ($(BUILD_TYPE),hipblas)
	ROCM_HOME ?= "/opt/rocm"
	CXX="$(ROCM_HOME)"/llvm/bin/clang++
	CC="$(ROCM_HOME)"/llvm/bin/clang
	GPU_TARGETS ?= gfx900,gfx90a,gfx1030,gfx1031,gfx1100
	AMDGPU_TARGETS ?= "$(GPU_TARGETS)"
	CMAKE_ARGS+=-DGGML_HIPBLAS=ON -DAMDGPU_TARGETS="$(AMDGPU_TARGETS)" -DGPU_TARGETS="$(GPU_TARGETS)"
	GGML_CUDA_OBJ_PATH=CMakeFiles/ggml-rocm.dir/ggml-cuda.cu.o
endif
ifeq ($(BUILD_TYPE),clblas)
	CMAKE_ARGS+=-DGGML_CLBLAST=ON
	CGO_TAGS=-tags cublas
endif
ifeq ($(BUILD_TYPE),metal)
	CMAKE_ARGS+=-DGGML_METAL=ON
	CGO_TAGS=-tags metal
	EXTRA_TARGETS+=ggml-metal
endif

ifdef CLBLAST_DIR
	CMAKE_ARGS+=-DCLBlast_dir=$(CLBLAST_DIR)
endif

#
# Print build information
#
$(info I chatglm.cpp build info: )
$(info I UNAME_S:  $(UNAME_S))
$(info I UNAME_P:  $(UNAME_P))
$(info I UNAME_M:  $(UNAME_M))
$(info I CXXFLAGS: $(CXXFLAGS))
$(info I BUILD_TYPE:  $(BUILD_TYPE))
$(info I CMAKE_ARGS:  $(CMAKE_ARGS))
$(info I EXTRA_TARGETS:  $(EXTRA_TARGETS))
$(info I CC:       $(CCV))
$(info I CXX:      $(CXXV))
$(info I CP:       $(CP))
$(info I CGO_TAGS:    $(CGO_TAGS))
$(info )

# Use this if you want to set the default behavior

prepare:
	mkdir -p build && mkdir -p out

# build chatglm.cpp
build/chatglm.cpp: prepare
	cd build && CC="$(CC)" CXX="$(CXX)" cmake $(CMAKE_ARGS) ../chatglm.cpp && VERBOSE=1 cmake --build . -j --config Release

# chatglm.dir
chatglm.dir: build/chatglm.cpp
	cd out && mkdir -p chatglm.dir
	$(CP) build$(DELIMITER)CMakeFiles$(DELIMITER)chatglm.dir$(DELIMITER)chatglm.cpp.o out$(DELIMITER)chatglm.dir$(DELIMITER)

# ggml.dir
ggml.dir: build/chatglm.cpp
	cd out && mkdir -p ggml.dir
	$(CP) build$(DELIMITER)third_party$(DELIMITER)ggml$(DELIMITER)src$(DELIMITER)CMakeFiles$(DELIMITER)ggml.dir$(DELIMITER)*.o out$(DELIMITER)ggml.dir$(DELIMITER)

# sentencepiece.dir
sentencepiece.dir: build/chatglm.cpp
	cd out && mkdir -p sentencepiece.dir
	$(CP) build$(DELIMITER)third_party$(DELIMITER)sentencepiece$(DELIMITER)src$(DELIMITER)CMakeFiles$(DELIMITER)sentencepiece-static.dir$(DELIMITER)*.cc.o out$(DELIMITER)sentencepiece.dir$(DELIMITER)
	$(CP) build$(DELIMITER)third_party$(DELIMITER)sentencepiece$(DELIMITER)src$(DELIMITER)CMakeFiles$(DELIMITER)sentencepiece-static.dir$(DELIMITER)builtin_pb$(DELIMITER)*.cc.o out$(DELIMITER)sentencepiece.dir$(DELIMITER)

# protobuf-lite.dir
protobuf-lite.dir: sentencepiece.dir
	cd out && mkdir -p protobuf-lite.dir
	$(CP) build$(DELIMITER)third_party$(DELIMITER)sentencepiece$(DELIMITER)src$(DELIMITER)CMakeFiles$(DELIMITER)sentencepiece-static.dir$(DELIMITER)__$(DELIMITER)third_party$(DELIMITER)protobuf-lite$(DELIMITER)*.cc.o out$(DELIMITER)protobuf-lite.dir$(DELIMITER)

# absl.dir
absl.dir: sentencepiece.dir
	cd out && mkdir -p absl.dir
	$(CP) build$(DELIMITER)third_party$(DELIMITER)sentencepiece$(DELIMITER)src$(DELIMITER)CMakeFiles$(DELIMITER)sentencepiece-static.dir$(DELIMITER)__$(DELIMITER)third_party$(DELIMITER)absl$(DELIMITER)flags$(DELIMITER)flag.cc.o out$(DELIMITER)absl.dir$(DELIMITER)

# ggml-metal
ggml-metal: ggml.dir
	$(CP) build/bin/ggml-metal.metal .

# binding
binding.o: prepare build/chatglm.cpp chatglm.dir ggml.dir sentencepiece.dir protobuf-lite.dir absl.dir
	$(CXX) $(CXXFLAGS) \
	-I./chatglm.cpp  \
	-I./chatglm.cpp/third_party/ggml/include/ggml \
	-I./chatglm.cpp/third_party/sentencepiece/src \
	binding.cpp -MD -MT binding.o -MF binding.d -o binding.o -c

libbinding.a: prepare binding.o $(EXTRA_TARGETS)
	ar src libbinding.a  \
	out/chatglm.dir/*.o \
	out/ggml.dir/*.o out/sentencepiece.dir/*.o  \
	out/protobuf-lite.dir/*.o out/absl.dir/*.o \
	binding.o

clean:
	rm -rf *.o
	rm -rf *.d
	rm -rf *.a
	rm -rf out
	rm -rf build

DOWNLOAD_TARGETS=ggllm-test-model.bin
ifeq ($(OS),Windows_NT)
	DOWNLOAD_TARGETS:=windows/ggllm-test-model.bin
endif

ggllm-test-model.bin:
	wget -q -N https://huggingface.co/Xorbits/chatglm3-6B-GGML/resolve/main/chatglm3-ggml-q4_0.bin -O ggllm-test-model.bin
windows/ggllm-test-model.bin:
	powershell -Command "Invoke-WebRequest -Uri 'https://huggingface.co/Xorbits/chatglm3-6B-GGML/resolve/main/chatglm3-ggml-q4_0.bin' -OutFile 'ggllm-test-model.bin'"

test: $(DOWNLOAD_TARGETS) libbinding.a
	go test ${CGO_TAGS} -timeout 1800s -o go-chatglm.cpp.test -c -cover
	TEST_MODEL=ggllm-test-model.bin ./go-chatglm.cpp.test
