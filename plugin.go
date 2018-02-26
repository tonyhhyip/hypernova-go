package hypernova

type Plugin interface {
	GetViewData(name string, data map[string]interface{}) (map[string]interface{}, error)
	PrepareRequest(jobs []*Job, originalJobs []*Job) []*Job
	ShouldSendRequest(jobs []*Job) bool
	WillSendRequest(jobs []*Job)
	OnError(err error, jobs []*Job)
	OnSuccess(job *JobResult)
	AfterResponse(results []*JobResult) []*JobResult
}
