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
#endif
#if defined (_WIN32)
#define WIN32_LEAN_AND_MEAN
#ifndef NOMINMAX
#define NOMINMAX
#endif
#include <io.h>
#include <stdio.h>
#include <windows.h>
#endif

#ifdef GGML_USE_CUBLAS
#include <ggml-cuda.h>
#endif

#ifdef GGML_USE_METAL
#include <ggml-metal.h>
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

std::vector<chatglm::ChatMessage> create_chat_message_vector(void** history, int count) {
    std::vector<chatglm::ChatMessage>* vec = new std::vector<chatglm::ChatMessage>;
    for (int i = 0; i < count; i++) {
        chatglm::ChatMessage* msg = (chatglm::ChatMessage*) history[i];
        vec->push_back(*msg);
    }

    return *vec;
}

std::vector<chatglm::ToolCallMessage> create_tool_call_vector(void** tool_calls, int count) {
    std::vector<chatglm::ToolCallMessage>* vec = new std::vector<chatglm::ToolCallMessage>;
    for (int i = 0; i < count; i++) {
        chatglm::ToolCallMessage* msg = (chatglm::ToolCallMessage*) tool_calls[i];
        vec->push_back(*msg);
    }

    return *vec;
}

std::string decode_with_special_tokens(chatglm::ChatGLM3Tokenizer* tokenizer, const std::vector<int> &ids) {
    std::vector<std::string> pieces;
    for (int id : ids) {
        auto pos = tokenizer->index_special_tokens.find(id);
        if (pos != tokenizer->index_special_tokens.end()) {
            // special tokens
            pieces.emplace_back(pos->second);
        } else {
            // normal tokens
            pieces.emplace_back(tokenizer->sp.IdToPiece(id));
        }
    }

    std::string text = tokenizer->sp.DecodePieces(pieces);
    return text;
}

void* load_model(const char *name) {
    return new chatglm::Pipeline(name);
}

int chat(void* pipe_pr, void** history, int history_count, void* params_ptr, char* result) {
    std::vector<chatglm::ChatMessage> vectors = create_chat_message_vector(history, history_count);
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    chatglm::ChatMessage res = pipe_p->chat(vectors, *params);

    std::string out = res.content;
    // ChatGLM3Tokenizer::decode_message change origin output, convert it to ChatMessage
    // So we need to convert it back
    if (pipe_p->model->config.model_type == chatglm::ModelType::CHATGLM3) {
        std::vector<chatglm::ChatMessage>* resultVec =  new std::vector<chatglm::ChatMessage>{res};
        chatglm::ChatGLM3Tokenizer* tokenizer = dynamic_cast<chatglm::ChatGLM3Tokenizer*>(pipe_p->tokenizer.get());
        std::vector<int> input_ids =  tokenizer->encode_messages(*resultVec, params->max_context_length);
        out = decode_with_special_tokens(tokenizer, input_ids);
    }
    strcpy(result, out.c_str());

    vectors.clear();
    return 0;
}

int stream_chat(void* pipe_pr, void** history, int history_count,void* params_ptr, char* result) {
    std::vector<chatglm::ChatMessage> vectors = create_chat_message_vector(history, history_count);
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;

    TextBindStreamer* text_stream = new TextBindStreamer(pipe_p->tokenizer.get(), pipe_pr);

    chatglm::ChatMessage res = pipe_p->chat(vectors, *params, text_stream);

    std::string out = res.content;
    // ChatGLM3Tokenizer::decode_message change origin output, convert it to ChatMessage
    // So we need to convert it back
    if (pipe_p->model->config.model_type == chatglm::ModelType::CHATGLM3) {
        std::vector<chatglm::ChatMessage>* resultVec =  new std::vector<chatglm::ChatMessage>{res};
        chatglm::ChatGLM3Tokenizer* tokenizer = dynamic_cast<chatglm::ChatGLM3Tokenizer*>(pipe_p->tokenizer.get());
        std::vector<int> input_ids =  tokenizer->encode_messages(*resultVec, params->max_context_length);
        out = decode_with_special_tokens(tokenizer, input_ids);
    }
    strcpy(result, out.c_str());

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

int get_embedding(void* pipe_pr, const char *prompt, int max_length, int * result) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;

    std::vector<int> embeddings = pipe_p->tokenizer->encode(prompt, max_length);

    for (size_t i = 0; i < embeddings.size(); i++) {
        result[i]=embeddings[i];
    }

    return 0;
}

void* allocate_params(int max_length, int max_context_length, bool do_sample, int top_k,
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

void free_params(void* params_ptr) {
    chatglm::GenerationConfig* params = (chatglm::GenerationConfig*) params_ptr;
    delete params;
}

void free_model(void* pipe_pr) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    delete pipe_p;
}

void* create_chat_message(const char* role, const char *content, void** tool_calls, int tool_calls_count) {
    std::vector<chatglm::ToolCallMessage> vector = create_tool_call_vector(tool_calls, tool_calls_count);
    return new chatglm::ChatMessage(role, content, vector);
}

void* create_tool_call(const char* type, void* codeOrFunc) {
    if (type == chatglm::ToolCallMessage::TYPE_FUNCTION) {
        chatglm::FunctionMessage* function_p = (chatglm::FunctionMessage*) codeOrFunc;
        return new chatglm::ToolCallMessage(*function_p);
    } else if (type == chatglm::ToolCallMessage::TYPE_CODE) {
        chatglm::CodeMessage* code_p = (chatglm::CodeMessage*) codeOrFunc;
        return new chatglm::ToolCallMessage(*code_p);
    }
    return nullptr;
}

void* create_function(const char* name, const char *arguments) {
    return new chatglm::FunctionMessage(name, arguments);
}


void* create_code(const char* input) {
    return  new chatglm::CodeMessage(input);;
}

char* get_model_type(void* pipe_pr) {
    chatglm::Pipeline* pipe_p = (chatglm::Pipeline*) pipe_pr;
    return strdup(to_string(pipe_p->model->config.model_type).data());
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
    if (!streamCallback(draft_pipe, printable_text.data())) {
        return;
    }
}

// copy from chatglm::TextStreamer
void TextBindStreamer::end() {
    std::string text = tokenizer_->decode(token_cache_);
    // callback go function
    if (!streamCallback(draft_pipe, text.substr(print_len_).data())) {
        return;
    }
    is_prompt_ = true;
    token_cache_.clear();
    print_len_ = 0;
}




