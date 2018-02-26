package hypernova

type Job struct {
	Name     string                 `json:"name"`
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}
