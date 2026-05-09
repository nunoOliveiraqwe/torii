package bus

type Topic string

const (
	//general topic
	TopicRequestBlocked   Topic = "request.blocked"
	TopicRequestProcessed Topic = "request.processed"
	//specific topics, this will be handy down the road
	TopicRateLimitTriggered  Topic = "request.rate_limited"
	TopicHoneypotTriggered   Topic = "honeypot.triggered"
	TopicIPFilterMatched     Topic = "ip.filter_matched"
	TopicCountryBlocked      Topic = "country.blocked"
	TopicUserAgentBlocked    Topic = "ua.blocked"
	TopicWAFBlocked          Topic = "waf.blocked"
	TopicHeaderPolicyBlocked Topic = "header.cmp.blocked"

	TopicBackendDown      Topic = "backend.down"
	TopicBackendRecovered Topic = "backend.recovered"

	//this will come when I build an IP rep system
	//event bus is step one
	TopicSuspiciousIP       Topic = "ip.suspicious"
	TopicIPScoreChanged     Topic = "ip.score_changed"
	TopicUARotationDetected Topic = "ua.rotation_detected"
	TopicAlertRaised        Topic = "alert.raised"
)
