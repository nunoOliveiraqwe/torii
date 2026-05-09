package bus

const (
	BlockActionDeny         = "deny"
	BlockActionTarpit       = "tarpit"
	BlockActionFakeResponse = "fake-response"
)

type RequestBlocked struct {
	Middleware string
	Reason     string
	StatusCode int
	RuleID     string
	RuleName   string
	Action     string
}

type RequestProcessed struct {
	ArrivedAt        int64
	ProcessingTimeMs int64
	StatusCode       int
	InternalId       string
	ConnectionName   string
	CountryCode      string
	ContinentCode    string
	UserAgent        string
	BytesSent        int64
	BytesReceived    int64
	Blocked          *RequestBlocked
}

type RequestRateLimited struct {
	Scope  string
	Key    string
	Limit  float64
	Burst  int
	Reason string
}

type HoneypotTriggered struct {
	Path           string
	MatchedPath    string
	Group          string
	CachedIPBefore bool
	Action         string
}

type IPFilterMatched struct {
	IP          string
	Source      string
	MatchedRule string
	List        string
}

type CountryBlocked struct {
	CountryCode   string
	ContinentCode string
	Reason        string
}

type UserAgentBlocked struct {
	UserAgent      string
	MatchedPattern string
	CachedIPBefore bool
}

type WAFBlocked struct {
	RuleID     string
	RuleName   string
	Message    string
	StatusCode int
	Tags       []string
}

type HeaderPolicyFailed struct {
	Header   string
	Rule     string
	Expected string
	Actual   string
}

type BackendDown struct {
	BackendURL          string
	Error               string
	ConsecutiveFailures int
}

type BackendRecovered struct {
	BackendURL string
}

type UARotationDetected struct {
	IP           string
	UserAgents   []string
	Window       string
	RequestCount int
}
