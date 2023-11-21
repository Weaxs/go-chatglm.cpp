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
#include <algorithm>

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

// stream for callback go function, copy from chatglm::TextStreamer
class TextBindStreamer : public chatglm::BaseStreamer {
public:
    TextBindStreamer(chatglm::BaseTokenizer *tokenizer, void* draft_pipe)
            : draft_pipe(draft_pipe), tokenizer_(tokenizer), is_prompt_(true), print_len_(0) {}
    void put(const std::vector<int> &output_ids) override;
    void end() override;

private:
    void* draft_pipe;
    chatglm::BaseTokenizer *tokenizer_;
    bool is_prompt_;
    std::vector<int> token_cache_;
    int print_len_;
};

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

int chat(void* pipe_pr, const char** history, int history_count, void* params_ptr, char* result) {
    std::vector<std::string> vectors = create_vector(history, history_count);
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    std::string res = pipe_p->chat(vectors, *params);
    strcpy(result, res.c_str());

    vectors.clear();

    return 0;
}

int stream_chat(void* pipe_pr, const char** history, int history_count,void* params_ptr, char* result) {
    std::vector<std::string> vectors = create_vector(history, history_count);
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    TextBindStreamer* text_stream = new TextBindStreamer(pipe_p->tokenizer.get(), pipe_pr);

    std::string res = pipe_p->chat(vectors, *params, text_stream);
    strcpy(result, res.c_str());

    vectors.clear();
    return 0;
}

int generate(void* pipe_pr, const char *prompt, void* params_ptr, char* result) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    std::string res = pipe_p->generate(std::string(prompt), *params);
    strcpy(result, res.c_str());

    return 0;
}

int stream_generate(void* pipe_pr, const char *prompt, void* params_ptr, char* result) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    TextBindStreamer* text_stream = new TextBindStreamer(pipe_p->tokenizer.get(), pipe_pr);

    std::string res = pipe_p->generate(std::string(prompt), *params, text_stream);
    strcpy(result, res.c_str());

    return 0;
}

int get_embedding(void* pipe_pr, void* params_ptr, const char *prompt, int * result) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    std::vector<int> embeddings = pipe_p->tokenizer->encode(prompt, params->max_length);

    for (size_t i = 0; i < embeddings.size(); i++) {
        result[i]=embeddings[i];
    }

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

// copy from chatglm::TextStreamer
void TextBindStreamer::put(const std::vector<int> &output_ids) {
    if (is_prompt_) {
        // skip prompt
        is_prompt_ = false;
        return;
    }

    static const std::vector<char> puncts{',', '!', ':', ';', '?'};

    token_cache_.insert(token_cache_.end(), output_ids.begin(), output_ids.end());
    std::string text = tokenizer_->decode(token_cache_);
    if (text.empty()) {
        return;
    }

    std::string printable_text;
    if (text.back() == '\n') {
        // flush the cache after newline
        printable_text = text.substr(print_len_);

        token_cache_.clear();
        print_len_ = 0;
    } else if (std::find(puncts.begin(), puncts.end(), text.back()) != puncts.end()) {
        // last symbol is a punctuation, hold on
    } else if (text.size() >= 3 && text.compare(text.size() - 3, 3, "�") == 0) {
        // ends with an incomplete token, hold on
    } else {
        printable_text = text.substr(print_len_);
        print_len_ = text.size();
    }

    // callback go function
    if (!streamCallback(draft_pipe, (char*)printable_text.c_str())) {
        return;
    }
}

// copy from chatglm::TextStreamer
void TextBindStreamer::end() {
    std::string text = tokenizer_->decode(token_cache_);
    // callback go function
    if (!streamCallback(draft_pipe, (char*)text.substr(print_len_).c_str())) {
        return;
    }
    is_prompt_ = true;
    token_cache_.clear();
    print_len_ = 0;
}




