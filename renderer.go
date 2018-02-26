package hypernova

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/satori/go.uuid"
)

var ErrEmptyResult = errors.New("server response missing results")

type Renderer struct {
	plugins     []Plugin
	Url         string
	config      map[string]interface{}
	incomingJob []*Job
	Client      *http.Client
}

func (r *Renderer) AddPlugin(plugin Plugin) {
	r.plugins = append(r.plugins, plugin)
}

func (r *Renderer) AddJob(id string, job *Job) {
	r.incomingJob = append(r.incomingJob, job)
}

func (r *Renderer) Render() *Response {
	jobs := r.createJobs()
	shouldSendRequest, jobs := r.prepareRequest(jobs)
	if !shouldSendRequest {
		return r.fallback(nil, jobs)
	}

	if resp, err := r.makeRequest(jobs); err != nil {
		return r.fallback(err, jobs)
	} else {
		return resp
	}
}

func (r *Renderer) createJobs() (jobs []*Job) {
	jobs = make([]*Job, 0)

	for _, job := range r.incomingJob {

		createdJob := &Job{
			Name:     job.Name,
			Data:     job.Data,
			Metadata: job.Metadata,
		}

		for _, plugin := range r.plugins {
			if data, err := plugin.GetViewData(createdJob.Name, createdJob.Data); err != nil {
				plugin.OnError(err, r.incomingJob)
			} else {
				createdJob.Data = data
			}
		}

		jobs = append(jobs, createdJob)
	}

	return
}

func (r *Renderer) prepareRequest(jobs []*Job) (shouldSend bool, preparedJobs []*Job) {
	preparedJobs = jobs
	shouldSend = false

	for _, plugin := range r.plugins {
		preparedJobs = plugin.PrepareRequest(preparedJobs, jobs)
	}

	for _, plugin := range r.plugins {
		if !plugin.ShouldSendRequest(preparedJobs) {
			return
		}
	}

	shouldSend = true
	return
}

func (r *Renderer) fallback(err error, jobs []*Job) (response *Response) {
	response = new(Response)

	response.Err = err
	response.Results = make([]*JobResult, 0)

	for _, job := range jobs {
		jobResult := new(JobResult)
		id := uuid.NewV4()
		jobResult.HTML = r.getFallbackHTML(job.Name, job.Data, id)
		jobResult.Meta = map[string]string{
			"uuid": id.String(),
		}
		jobResult.OriginalJob = job
		response.Results = append(response.Results, jobResult)
	}

	return
}

func (r *Renderer) getFallbackHTML(moduleName string, data map[string]interface{}, uuid uuid.UUID) string {
	content, _ := json.Marshal(data)
	return fmt.Sprintf(
		`<div data-hypernova-key="%s" data-hypernova-id="%s"></div>
		<script type="application/json" data-hypernova-key="%s" data-hypernova-id="%s"><!--%s--></script>`,
		moduleName,
		uuid,
		moduleName,
		uuid,
		string(content),
	)
}

func (r *Renderer) makeRequest(jobs []*Job) (*Response, error) {
	for _, plugin := range r.plugins {
		plugin.WillSendRequest(jobs)
	}

	results, err := r.doRequest(jobs)
	if err != nil {
		return nil, err
	}
	return r.finalize(results), nil
}

func (r *Renderer) doRequest(jobs []*Job) (results []*JobResult, err error) {
	content, err := json.Marshal(jobs)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", r.Url, bytes.NewBuffer(content))
	if err != nil {
		return
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var reply struct {
		Result []struct {
			HTML     string            `json:"html"`
			Err      string            `json:"error,omitempty"`
			Success  bool              `json:"success"`
			Meta     map[string]string `json:"meta,omitempty"`
			Duration float64           `json:"duration"`
		} `json:"result"`
		Error string `json:"error,omitempty"`
	}

	if err = json.Unmarshal(body, &reply); err != nil {
		return
	}

	if len(reply.Result) == 0 {
		return nil, ErrEmptyResult
	}

	if reply.Error != "" {
		for _, plugin := range r.plugins {
			plugin.OnError(errors.New(reply.Error), jobs)
		}
	}

	results = []*JobResult{}
	for key, result := range reply.Result {
		var e error = nil
		if result.Err != "" {
			e = errors.New(result.Err)
		}
		results[key] = &JobResult{
			Err:         e,
			HTML:        result.HTML,
			Success:     result.Success,
			Meta:        result.Meta,
			OriginalJob: jobs[key],
			Duration:    time.Duration(result.Duration * float64(time.Second)),
		}
	}

	return
}

func (r *Renderer) finalize(results []*JobResult) *Response {
	for _, result := range results {
		if result.Err != nil {
			for _, plugin := range r.plugins {
				plugin.OnError(result.Err, []*Job{result.OriginalJob})
			}
		}
	}

	for _, result := range results {
		if result.Success {
			for _, plugin := range r.plugins {
				plugin.OnSuccess(result)
			}
		}
	}

	for _, plugin := range r.plugins {
		results = plugin.AfterResponse(results)
	}

	return &Response{
		Results: results,
	}
}
