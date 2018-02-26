package hypernova

type Plugin interface {
	GetViewData(name string, data map[string]interface{}) (map[string]interface{}, error)
	PrepareRequest(jobs Jobs, originalJobs Jobs) Jobs
	ShouldSendRequest(jobs Jobs) bool
	WillSendRequest(jobs Jobs)
	OnError(err error, jobs Jobs)
	OnSuccess(job *JobResult)
	AfterResponse(results JobResults) JobResults
}
