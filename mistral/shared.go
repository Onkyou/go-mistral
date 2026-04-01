package mistral

type UsageInfo struct {
	CompletionTokens    int                  `json:"completion_tokens"`
	NumCachedTokens     *int                 `json:"num_cached_tokens,omitempty"`
	PromptAudioSeconds  *int                 `json:"prompt_audio_seconds,omitempty"`
	PromptTokenDetails  *PromptTokenDetails  `json:"prompt_token_details,omitempty"`
	PromptTokens        int                  `json:"prompt_tokens"`
	PromptTokensDetails *PromptTokensDetails `json:"prompt_tokens_details,omitempty"`
	TotalTokens         int                  `json:"total_tokens"`
}

type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}
