package constant

import "time"

var CodePefix = map[string]string{
	"client":          "CL",
	"company":         "CO",
	"group":           "GR",
	"mechant":         "ME",
	"principle":       "PR",
	"role":            "RO",
	"profile":         "PF",
	"acquirer":        "AC",
	"bin":             "BI",
	"stock":           "SL",
	"product":         "ST",
	"user":            "US",
	"issuer":          "IS",
	"main_feature":    "FT",
	"support_feature": "HS",
}

const (
	IconStatusLoading string = "loading"
	IconStatusFailed  string = "failed"
)

const (
	EnvironmentLocal       string = "local"
	EnvironmentDevelopment string = "development"
	EnvironmentStaging     string = "staging"
	EnvironmentSandbox     string = "sandbox"
	EnvironmentProduction  string = "production"
)

type contextKey string

const (
	CtxKeySubLogger       string     = "sub_logger"
	CtxKeyAuthPayload     string     = "auth_payload"
	CtxKeyRequestID       string     = "request_id"
	CtxKeyTraceID         string     = "trace_id"
	CtxKeyLoggerStartTime contextKey = "redis_logger_start_time"
	CtxKeyRequestIP       contextKey = "request_ip"
)

const (
	TokenType          string = "Bearer"
	TokenMinSecretSize int    = 32
	TokenIssuer        string = "goapptemp-auth"
)

const (
	ImgMaxSize int    = 10 * 1024 * 1024
	FailedIcon string = "failed"
)

const (
	SoftDeleteColumnName string = "deleted_at"
	KeyColumnName        string = "key"
	ParentSchema         string = "goapptemp"
)

const (
	ClientModelType string = "client"
)

const (
	IpRateLimitAttempts     int           = 50
	IpRateLimitWindow       time.Duration = 10 * time.Minute
	IpBackoffBaseSeconds    int           = 60
	UserFailedAttemptsLimit int           = 10
	UserFailedWindow        time.Duration = 15 * time.Minute
	UserLockoutDuration     time.Duration = 30 * time.Minute
)

const (
	DummyPasswordHash string = "$2a$12$7bJc1sFjQw2u3Yv9kqLq.eK1gY4Gx8h0q1zQj2Vl9cYQ1N0e8xF5e"
)

const (
	MinPasswordLength = 8
)
