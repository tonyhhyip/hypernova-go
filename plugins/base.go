package plugins

import "github.com/tonyhhyip/hypernova-go"

var _ hypernova.Plugin = &PluginBase{}

type PluginBase struct{}

func (*PluginBase) GetViewData(_ string, data map[string]interface{}) (map[string]interface{}, error) {
	return data, nil
}

func (*PluginBase) PrepareRequest(jobs []*hypernova.Job, _ []*hypernova.Job) []*hypernova.Job {
	return jobs
}

func (*PluginBase) ShouldSendRequest(jobs []*hypernova.Job) bool {
	return true
}

func (*PluginBase) WillSendRequest(jobs []*hypernova.Job) {}

func (*PluginBase) OnError(err error, jobs []*hypernova.Job) {}

func (*PluginBase) OnSuccess(job *hypernova.JobResult) {}

func (*PluginBase) AfterResponse(results []*hypernova.JobResult) []*hypernova.JobResult {
	return results
}
