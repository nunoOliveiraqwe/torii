package configuration

type Middleware struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"-"`
}
