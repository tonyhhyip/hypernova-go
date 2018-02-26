package plugins

import "github.com/tonyhhyip/hypernova-go"

var _ hypernova.Plugin = &PluginBase{}

type PluginBase struct{}

func (*PluginBase) GetViewData(_ string, data map[string]interface{}) (map[string]interface{}, error) {
	return data, nil
}

func (*PluginBase) PrepareRequest(jobs hypernova.Jobs, originalJobs hypernova.Jobs) hypernova.Jobs {
	return jobs
}

func (*PluginBase) ShouldSendRequest(jobs hypernova.Jobs) bool {
	return true
}

func (*PluginBase) WillSendRequest(jobs hypernova.Jobs) {}

func (*PluginBase) OnError(err error, jobs hypernova.Jobs) {}

func (*PluginBase) OnSuccess(job *hypernova.JobResult) {}

func (*PluginBase) AfterResponse(results hypernova.JobResults) hypernova.JobResults {
	return results
}
