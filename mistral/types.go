package mistral

type Model string

const (
	ModelMistralLargeLatest    Model = "mistral-large-latest"
	ModelMistralMediumLatest   Model = "mistral-medium-latest"
	ModelMistralSmallLatest    Model = "mistral-small-latest"
	ModelCodestralLatest       Model = "codestral-latest"
	ModelOpenMixtral8x7b       Model = "open-mixtral-8x7b"
	ModelOpenMixtral8x22b      Model = "open-mixtral-8x22b"
	ModelOpenMistral7b         Model = "open-mistral-7b"
	ModelMistralLarge2402      Model = "mistral-large-2402"
	ModelMistralMedium2312     Model = "mistral-medium-2312"
	ModelMistralSmall2402      Model = "mistral-small-2402"
	ModelMistralSmall2312      Model = "mistral-small-2312"
	ModelMistralTiny           Model = "mistral-tiny-2312"
	ModelMistralEmbed          Model = "mistral-embed"
	ModelMistralModeration2603 Model = "mistral-moderation-2603"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type ReasoningEffort string

const (
	ReasoningEffortHigh   ReasoningEffort = "high"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortNone   ReasoningEffort = "none"
)
