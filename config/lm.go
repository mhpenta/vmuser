package config

type LLM struct {
	OpenAIKey    string `toml:"OpenAIAPIKey"`
	AnthropicKey string `toml:"AnthropicAPIKey"`
	GeminiKey    string `toml:"GeminiAPIKey"`
}

type LLMLibConfig struct {
	LLMLibURL    string `toml:"LLMLibURL"`
	LLMLibAPIKey string `toml:"LLMLibAPIKey"`
}
