#include "chatglm.h"

#include "binding.h"
#include <string>
#include <vector>
#include <sstream>
#include <iostream>
#include <cassert>
#include <cinttypes>
#include <cmath>
#include <cstdio>
#include <cstring>
#include <fstream>

#if defined (__unix__) || (defined (__APPLE__) && defined (__MACH__))
#include <signal.h>
#include <unistd.h>
#elif defined (_WIN32)
#define WIN32_LEAN_AND_MEAN
#define NOMINMAX
#include <windows.h>
#include <signal.h>
#endif

#if defined (__unix__) || (defined (__APPLE__) && defined (__MACH__)) || defined (_WIN32)
void sigint_handler(int signo) {
    if (signo == SIGINT) {
            _exit(130);
    }
}
#endif

std::vector<std::string> create_vector(const char** strings, int count) {
    auto vec = new std::vector<std::string>;
    for (int i = 0; i < count; i++) {
      vec->push_back(std::string(strings[i]));
    }
    return *vec;
}

void* load_model(const char *name) {
    return new chatglm::Pipeline(name);
}

int chat(void* pipe_pr, const char** history, int history_count, void* params_ptr, bool stream, char* result) {
    std::vector<std::string> vectors = create_vector(history, history_count);
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    // TODO: support stream
    std::string res = pipe_p->chat(vectors, *params);
    strcpy(result, res.c_str());

    vectors.clear();

    return 0;
}

int generate(void* pipe_pr, const char *prompt, void* params_ptr, bool stream, char* result) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    // TODO: support streamer
//    chatglm::PerfStreamer *streamer = nullptr;
//    if (stream) {
//        streamer = new chatglm::PerfStreamer;
//    }
    std::string res = pipe_p->generate(std::string(prompt), *params);
    strcpy(result, res.c_str());

//    if (streamer != nullptr) {
//        delete streamer;
//    }

    return 0;
}

void* chatglm_allocate_params(int max_length, int max_context_length, bool do_sample, int top_k,
                                float top_p, float temperature, float repetition_penalty, int num_threads) {
    chatglm::GenerationConfig* gen_config = new chatglm::GenerationConfig;
    gen_config->max_length = max_length;
    gen_config->max_context_length = max_context_length;
    gen_config->do_sample = do_sample;
    gen_config->top_k = top_k;
    gen_config->top_p = top_p;
    gen_config->temperature = temperature;
    gen_config->repetition_penalty = repetition_penalty;
    gen_config->num_threads = num_threads;
    return gen_config;
}

void chatglm_free_params(void* params_ptr) {
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;
    delete params;
}

void chatglm_free_model(void* pipe_pr) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    delete pipe_p;
}



