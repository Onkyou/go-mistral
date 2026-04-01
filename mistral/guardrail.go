package mistral

// GuardrailConfig represents the guardrail configuration. Struct is simplified to only support v2.
type GuardrailConfig struct {
	BlockOnError bool              `json:"block_on_error"`
	Moderation   *ModerationConfig `json:"moderation_llm_v2,omitempty"`
}

type ModerationConfigAction string

const (
	ModerationConfigActionBlock ModerationConfigAction = "block"
	ModerationConfigActionNone  ModerationConfigAction = "none"
)

type ModerationConfig struct {
	Action                   ModerationConfigAction        `json:"action"`
	IgnoreOtherCategories    bool                          `json:"ignore_other_categories"`
	ModelName                Model                         `json:"model_name"`
	CustomCategoryThresholds *ModerationCategoryThresholds `json:"custom_category_thresholds,omitempty"`
}

type ModerationCategoryThresholds struct {
	Criminal              float64 `json:"criminal,omitempty"`
	Dangerous             float64 `json:"dangerous,omitempty"`
	Financial             float64 `json:"financial,omitempty"`
	HateAndDiscrimination float64 `json:"hate_and_discrimination,omitempty"`
	Health                float64 `json:"health,omitempty"`
	Jailbreaking          float64 `json:"jailbreaking,omitempty"`
	Law                   float64 `json:"law,omitempty"`
	Pii                   float64 `json:"pii,omitempty"`
	Selfharm              float64 `json:"selfharm,omitempty"`
	Sexual                float64 `json:"sexual,omitempty"`
	ViolenceAndThreats    float64 `json:"violence_and_threats,omitempty"`
}
