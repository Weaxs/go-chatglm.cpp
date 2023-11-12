INCLUDE_PATH := $(abspath ./)
LIBRARY_PATH := $(abspath ./)

ifndef UNAME_S
UNAME_S := $(shell uname -s)
endif

ifndef UNAME_P
UNAME_P := $(shell uname -p)
endif

ifndef UNAME_M
UNAME_M := $(shell uname -m)
endif

CCV := $(shell $(CC) --version | head -n 1)
CXXV := $(shell $(CXX) --version | head -n 1)

# Mac OS + Arm can report x86_64
# ref: https://github.com/ggerganov/whisper.cpp/issues/66#issuecomment-1282546789
ifeq ($(UNAME_S),Darwin)
	ifneq ($(UNAME_P),arm))
		SYSCTL_M := $(shell sysctl -n hw.optional.arm64 2>/dev/null)
		ifeq ($(SYSCTL_M),1)
			# UNAME_P := arm
			# UNAME_M := arm64
			warn := $(warning Your arch is announced as x86_64, but it seems to actually be ARM64. Not fixing that can lead to bad performance. For more info see: https://github.com/ggerganov/whisper.cpp/issues/66\#issuecomment-1282546789)
		endif
	endif
endif

#
# Compile flags
#

BUILD_TYPE?=
# keep standard at C17 and C++17
CFLAGS   = -I./chatglm.cpp -I. -O3 -DNDEBUG -std=c17 -fPIC
CXXFLAGS = -I./chatglm.cpp -I. -I./chatglm.cpp/third_party -I./third_party -O3 -DNDEBUG -std=c++17 -fPIC
LDFLAGS  =

# warnings
CFLAGS   += -Wall -Wextra -Wpedantic -Wcast-qual -Wdouble-promotion -Wshadow -Wstrict-prototypes -Wpointer-arith -Wno-unused-function
CXXFLAGS += -Wall -Wextra -Wpedantic -Wcast-qual -Wno-unused-function

# OS specific
# TODO: support Windows
ifeq ($(UNAME_S),Linux)
	CFLAGS   += -pthread
	CXXFLAGS += -pthread
endif
ifeq ($(UNAME_S),Darwin)
	CFLAGS   += -pthread
	CXXFLAGS += -pthread
endif
ifeq ($(UNAME_S),FreeBSD)
	CFLAGS   += -pthread
	CXXFLAGS += -pthread
endif
ifeq ($(UNAME_S),NetBSD)
	CFLAGS   += -pthread
	CXXFLAGS += -pthread
endif
ifeq ($(UNAME_S),OpenBSD)
	CFLAGS   += -pthread
	CXXFLAGS += -pthread
endif
ifeq ($(UNAME_S),Haiku)
	CFLAGS   += -pthread
	CXXFLAGS += -pthread
endif

# GPGPU specific
GGML_CUDA_OBJ_PATH=third_party/ggml/src/CMakeFiles/ggml.dir/ggml-cuda.cu.o


# Architecture specific
# TODO: probably these flags need to be tweaked on some architectures
#       feel free to update the Makefile for your architecture and send a pull request or issue
ifeq ($(UNAME_M),$(filter $(UNAME_M),x86_64 i686))
	# Use all CPU extensions that are available:
	CFLAGS += -march=native -mtune=native
endif
ifneq ($(filter ppc64%,$(UNAME_M)),)
	POWER9_M := $(shell grep "POWER9" /proc/cpuinfo)
	ifneq (,$(findstring POWER9,$(POWER9_M)))
		CFLAGS += -mcpu=power9
		CXXFLAGS += -mcpu=power9
	endif
	# Require c++23's std::byteswap for big-endian support.
	ifeq ($(UNAME_M),ppc64)
		CXXFLAGS += -std=c++23 -DGGML_BIG_ENDIAN
	endif
endif
ifndef CHATGLM_NO_ACCELERATE
	# Mac M1 - include Accelerate framework.
	# `-framework Accelerate` works on Mac Intel as well, with negliable performance boost (as of the predict time).
	ifeq ($(UNAME_S),Darwin)
		CFLAGS  += -DGGML_USE_ACCELERATE
		LDFLAGS += -framework Accelerate
	endif
endif
ifdef CHATGLM_OPENBLAS
	CFLAGS  += -DGGML_USE_OPENBLAS -I/usr/local/include/openblas
	LDFLAGS += -lopenblas
endif
ifdef CHATGLM_GPROF
	CFLAGS   += -pg
	CXXFLAGS += -pg
endif
ifneq ($(filter aarch64%,$(UNAME_M)),)
	CFLAGS += -mcpu=native
	CXXFLAGS += -mcpu=native
endif
ifneq ($(filter armv6%,$(UNAME_M)),)
	# Raspberry Pi 1, 2, 3
	CFLAGS += -mfpu=neon-fp-armv8 -mfp16-format=ieee -mno-unaligned-access
endif
ifneq ($(filter armv7%,$(UNAME_M)),)
	# Raspberry Pi 4
	CFLAGS += -mfpu=neon-fp-armv8 -mfp16-format=ieee -mno-unaligned-access -funsafe-math-optimizations
endif
ifneq ($(filter armv8%,$(UNAME_M)),)
	# Raspberry Pi 4
	CFLAGS += -mfp16-format=ieee -mno-unaligned-access
endif

ifeq ($(BUILD_TYPE),openblas)
	EXTRA_LIBS=
	CMAKE_ARGS+=-DCHATGLM_BLAS=ON -DCHATGLM_BLAS_VENDOR=OpenBLAS -DBLAS_INCLUDE_DIRS=/usr/include/openblas
endif

#ifeq ($(BUILD_TYPE),blis)
#	EXTRA_LIBS=
#	CMAKE_ARGS+=-DCHATGLM_BLAS=ON -DCHATGLM_BLAS_VENDOR=FLAME
#endif

ifeq ($(BUILD_TYPE),cublas)
	EXTRA_LIBS=
	CMAKE_ARGS+=-DCHATGLM_CUBLAS=ON
	EXTRA_TARGETS+=chatglm.cpp/ggml-cuda.o
endif

ifeq ($(BUILD_TYPE),hipblas)
	ROCM_HOME ?= "/opt/rocm"
	CXX="$(ROCM_HOME)"/llvm/bin/clang++
	CC="$(ROCM_HOME)"/llvm/bin/clang
	EXTRA_LIBS=
	GPU_TARGETS ?= gfx900,gfx90a,gfx1030,gfx1031,gfx1100
	AMDGPU_TARGETS ?= "$(GPU_TARGETS)"
	CMAKE_ARGS+=-DCHATGLM_HIPBLAS=ON -DAMDGPU_TARGETS="$(AMDGPU_TARGETS)" -DGPU_TARGETS="$(GPU_TARGETS)"
	EXTRA_TARGETS+=chatglm.cpp/ggml-cuda.o
	GGML_CUDA_OBJ_PATH=CMakeFiles/ggml-rocm.dir/ggml-cuda.cu.o
endif

#ifeq ($(BUILD_TYPE),clblas)
#	EXTRA_LIBS=
#	CMAKE_ARGS+=-DCHATGLM_CLBLAST=ON
#	EXTRA_TARGETS+=chatglm.cpp/third_party/ggml-opencl.o
#endif

ifeq ($(BUILD_TYPE),metal)
	EXTRA_LIBS=
	CGO_LDFLAGS+="-framework Accelerate -framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders"
	CMAKE_ARGS+=-DCHATGLM_METAL=ON
	EXTRA_TARGETS+=chatglm.cpp/ggml-metal.o
endif

ifdef CLBLAST_DIR
	CMAKE_ARGS+=-DCLBlast_dir=$(CLBLAST_DIR)
endif

# TODO: support Windows
ifeq ($(GPU_TESTS),true)
	CGO_LDFLAGS="-lcublas -lcudart -L/usr/local/cuda/lib64/"
	TEST_LABEL=gpu
else
	TEST_LABEL=!gpu
endif

#
# Print build information
#

$(info I chatglm.cpp build info: )
$(info I UNAME_S:  $(UNAME_S))
$(info I UNAME_P:  $(UNAME_P))
$(info I UNAME_M:  $(UNAME_M))
$(info I CFLAGS:   $(CFLAGS))
$(info I CXXFLAGS: $(CXXFLAGS))
$(info I CGO_LDFLAGS:  $(CGO_LDFLAGS))
$(info I LDFLAGS:  $(LDFLAGS))
$(info I BUILD_TYPE:  $(BUILD_TYPE))
$(info I CMAKE_ARGS:  $(CMAKE_ARGS))
$(info I EXTRA_TARGETS:  $(EXTRA_TARGETS))
$(info I CC:       $(CCV))
$(info I CXX:      $(CXXV))
$(info )

# Use this if you want to set the default behavior

prepare:
	mkdir -p build && mkdir -p out

# build chatglm.cpp
build/chatglm.cpp: prepare
	cd build && CC="$(CC)" CXX="$(CXX)" cmake ../chatglm.cpp $(CMAKE_ARGS) && VERBOSE=1 cmake --build . -j --config Release

# chatglm.dir
chatglm.dir: build/chatglm.cpp
	cd out && mkdir -p chatglm.dir && cd ../build && \
	cp -rp CMakeFiles/chatglm.dir/chatglm.cpp.o ../out/chatglm.dir/chatglm.o

# ggml.dir
ggml.dir: build/chatglm.cpp
	cd out && mkdir -p ggml.dir && cd ../build && \
	cp -rf third_party/ggml/src/CMakeFiles/ggml.dir/ggml.c.o ../out/ggml.dir/ggml.o && \
	cp -rf third_party/ggml/src/CMakeFiles/ggml.dir/ggml-alloc.c.o ../out/ggml.dir/ggml-alloc.o

# sentencepiece.dir
sentencepiece.dir: build/chatglm.cpp
	cd out && mkdir -p sentencepiece.dir && cd ../build && \
	cp -rf third_party/sentencepiece/src/CMakeFiles/sentencepiece-static.dir/sentencepiece_processor.cc.o ../out/sentencepiece.dir/sentencepiece_processor.o && \
	cp -rf third_party/sentencepiece/src/CMakeFiles/sentencepiece-static.dir/error.cc.o ../out/sentencepiece.dir/error.o && \
	cp -rf third_party/sentencepiece/src/CMakeFiles/sentencepiece-static.dir/model_interface.cc.o ../out/sentencepiece.dir/model_interface.o

# protobuf-lite.dir
protobuf-lite.dir: sentencepiece.dir
	cd out && mkdir -p protobuf-lite.dir && cd ../build && \
	find third_party/sentencepiece/src/CMakeFiles/sentencepiece-static.dir/__/third_party/protobuf-lite -name '*.cc.o' -exec cp {} ../out/protobuf-lite.dir/ \;

# absl.dir
absl.dir: sentencepiece.dir
	cd out && mkdir -p absl.dir && cd ../build && \
	find third_party/sentencepiece/src/CMakeFiles/sentencepiece-static.dir/__/third_party/absl/flags/ -name '*.cc.o' -exec cp {} ../out/absl.dir/ \;

# binding
binding.o: prepare build/chatglm.cpp chatglm.dir ggml.dir sentencepiece.dir protobuf-lite.dir absl.dir
	$(CXX) $(CXXFLAGS) \
	-I./chatglm.cpp  \
	-I./chatglm.cpp/third_party/ggml/include/ggml \
	-I./chatglm.cpp/third_party/sentencepiece/src \
	binding.cpp -o binding.o -c $(LDFLAGS)

# ggml-cuda
ggml.dir/ggml-cuda.o: ggml.dir
	cd build && cp -rf "$(GGML_CUDA_OBJ_PATH)" ../out/ggml.dir/ggml-cuda.o

# ggml-opencl
ggml.dir/ggml-opencl.o: ggml.dir
	cd build && cp -rf third_party/ggml/src/CMakeFiles/ggml.dir/ggml-opencl.cpp.o ../out/ggml.dir/ggml-opencl.o

# ggml-metal
ggml.dir/ggml-metal.o: ggml.dir
	cd build && cp -rf third_party/ggml/src/CMakeFiles/ggml.dir/ggml-metal.m.o ../out/ggml.dir/ggml-metal.o

libbinding.a: prepare binding.o $(EXTRA_TARGETS)
	ar src libbinding.a  \
	out/chatglm.dir/chatglm.o \
	out/ggml.dir/*.o out/sentencepiece.dir/*.o  \
	out/protobuf-lite.dir/*.o out/absl.dir/*.o \
	binding.o

clean:
	rm -rf *.o
	rm -rf *.a
	rm -rf build


ggllm-test-model.bin:
	wget -q https://huggingface.co/Xorbits/chatglm3-6B-GGML/resolve/main/chatglm3-ggml-q4_0.bin -O ggllm-test-model.bin

