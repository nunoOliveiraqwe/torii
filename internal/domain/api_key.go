package domain

import "time"

type Scope string

const (
	READ_STATS_SCOPE   = Scope("read_stats")
	READ_CONFIG_SCOPE  = Scope("read_config")
	WRITE_CONFIG_SCOPE = Scope("write_config")
)

var AvailableScopesMap = map[Scope]byte{
	READ_STATS_SCOPE:   1 << 0,
	READ_CONFIG_SCOPE:  1 << 1,
	WRITE_CONFIG_SCOPE: 1 << 2,
}

type ApiKey struct {
	ID        int
	Alias     string         `json:"alias"`
	Key       string         `json:"key"`
	Scopes    map[Scope]byte `json:"scopes"` //it's just faster than to impl a hashset type, which will just be a wrapper around a map anyway, db then translates this into csv
	Expires   time.Time      `json:"expires"`
	CreatedAt int64          `json:"created_at"`
}
