package chatglm

type ModelOptions struct {
}

type GenerationOptions struct {
	MaxLength         int
	MaxContextLength  int
	DoSample          bool
	TopK              int
	TopP              float32
	Temperature       float32
	RepetitionPenalty float32
	NumThreads        int
	Stream            bool
}

type GenerationOption func(g *GenerationOptions)

var DefaultGenerationOptions GenerationOptions = GenerationOptions{
	MaxLength:         2048,
	MaxContextLength:  512,
	DoSample:          true,
	TopK:              0,
	TopP:              0.7,
	Temperature:       0.95,
	RepetitionPenalty: 1.0,
	NumThreads:        0,
	Stream:            false,
}

func NewGenerationOptions(opts ...GenerationOption) GenerationOptions {
	p := DefaultGenerationOptions
	for _, opt := range opts {
		opt(&p)
	}
	return p
}
