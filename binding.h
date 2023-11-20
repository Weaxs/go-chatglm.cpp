#ifdef __cplusplus
#include <string>
extern "C" {
#endif

#include <stdbool.h>

extern unsigned char tokenCallback(void *, char *);

void* load_model(const char *name);

int chat(void* pipe_pr, const char** history, int history_count, void* params_ptr, char* result);

void* stream_chat(void* pipe_pr, const char** history, int history_count, void* params_ptr);

int generate(void* pipe_pr, const char *prompt, void* params_ptr, char* result);

void* stream_generate(void* pipe_pr, const char *prompt, void* params_ptr);

int stream_to_string(void* steamer, char* result);

int get_embedding(void* pipe_pr, void* params_ptr, const char *prompt, int * result);

void* chatglm_allocate_params(int max_length, int max_context_length, bool do_sample, int top_k,
                                float top_p, float temperature, float repetition_penalty, int num_threads);

void chatglm_free_params(void* params_ptr);

void chatglm_free_model(void* pipe_pr);

#ifdef __cplusplus
}

#endif