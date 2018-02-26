package hypernova

import "time"

type JobResult struct {
	HTML        string            `json:"html"`
	Err         error             `json:"error"`
	Success     bool              `json:"success"`
	OriginalJob *Job              `json:"-"`
	Meta        map[string]string `json:"meta,omitempty"`
	Duration    time.Duration     `json:"-"`
}

func (jr *JobResult) String() string {
	return jr.HTML
}
