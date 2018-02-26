package hypernova

type Response struct {
	Err     error
	Results []*JobResult
}
