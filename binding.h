#ifdef __cplusplus
#include <string>
extern "C" {
#endif

#include <stdbool.h>

extern bool streamCallback(void *, char *);

void* load_model(const char *name);

int chat(void* pipe_pr, void** history, int history_count, void* params_ptr, char* result);

int stream_chat(void* pipe_pr, void** history, int history_count, void* params_ptr, char* result);

int generate(void* pipe_pr, const char *prompt, void* params_ptr, char* result);

int stream_generate(void* pipe_pr, const char *prompt, void* params_ptr, char* result);

int get_embedding(void* pipe_pr, const char *prompt, int max_length, int * result);

void* allocate_params(int max_length, int max_context_length, bool do_sample, int top_k,
                      float top_p, float temperature, float repetition_penalty, int num_threads);

void free_params(void* params_ptr);

void free_model(void* pipe_pr);

void* create_chat_message(const char* role, const char *content, void** tool_calls, int tool_calls_count);

void* create_tool_call(const char* type, void* codeOrFunc);

void* create_function(const char* name, const char *arguments);

void* create_code(const char* code);

char* get_model_type(void* pipe_pr);

#ifdef __cplusplus
}



#endif