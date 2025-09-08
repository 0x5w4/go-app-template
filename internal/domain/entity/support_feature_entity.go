package entity

type SupportFeature struct {
	Base
	Code     string
	Name     string
	Key      string
	IsActive bool
}

type ValidatableString struct {
	Value   string `json:"value" validate:"required,min=2,max=50,alpha_space"`
	Message string `json:"message,omitempty"`
}

type ValidatableKey struct {
	Value   string `json:"value" validate:"required,min=2,max=50,username_chars_allowed"`
	Message string `json:"message,omitempty"`
}

type ValidatableBool struct {
	Value   *bool  `json:"value" validate:"required,boolean"`
	Message string `json:"message,omitempty"`
}

type SupportFeaturePreview struct {
	Row      int               `json:"row"`
	Name     ValidatableString `json:"name" validate:"required"`
	Key      ValidatableKey    `json:"key" validate:"required"`
	IsActive ValidatableBool   `json:"is_active" validate:"required"`
}
