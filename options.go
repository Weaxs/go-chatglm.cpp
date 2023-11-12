package chatglm

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

var DefaultGenerationOptions *GenerationOptions = &GenerationOptions{
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

func NewGenerationOptions(opts ...GenerationOption) *GenerationOptions {
	p := DefaultGenerationOptions
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func SetMaxLength(maxLength int) GenerationOption {
	return func(g *GenerationOptions) {
		g.MaxLength = maxLength
	}
}

func SetMaxContextLength(maxContextLength int) GenerationOption {
	return func(g *GenerationOptions) {
		g.MaxContextLength = maxContextLength
	}
}

func SetDoSample(doSample bool) GenerationOption {
	return func(g *GenerationOptions) {
		g.DoSample = doSample
	}
}

func SetTopK(topK int) GenerationOption {
	return func(g *GenerationOptions) {
		g.TopK = topK
	}
}

func SetTopP(topP float32) GenerationOption {
	return func(g *GenerationOptions) {
		g.TopP = topP
	}
}

func SetTemperature(temperature float32) GenerationOption {
	return func(g *GenerationOptions) {
		g.Temperature = temperature
	}
}

func SetRepetitionPenalty(repetitionPenalty float32) GenerationOption {
	return func(g *GenerationOptions) {
		g.RepetitionPenalty = repetitionPenalty
	}
}

func SetNumThreads(numThreads int) GenerationOption {
	return func(g *GenerationOptions) {
		g.NumThreads = numThreads
	}
}

func SetStream(stream bool) GenerationOption {
	return func(g *GenerationOptions) {
		g.Stream = stream
	}
}
