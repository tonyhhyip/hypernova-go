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
	incomingJob Jobs
	Client      *http.Client
}

func NewRenderer(url string) *Renderer {
	return &Renderer{
		plugins:     make([]Plugin, 0),
		Url:         url,
		incomingJob: make(Jobs),
		Client:      http.DefaultClient,
	}
}

// Add a plugin
func (r *Renderer) AddPlugin(plugin Plugin) {
	r.plugins = append(r.plugins, plugin)
}

// Add a job
func (r *Renderer) AddJob(id string, job *Job) {
	r.incomingJob[id] = job
}

// Render Page Component
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

func (r *Renderer) createJobs() (jobs Jobs) {
	jobs = make(Jobs)

	for name, job := range r.incomingJob {

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

		jobs[name] = createdJob
	}

	return
}

func (r *Renderer) prepareRequest(jobs Jobs) (shouldSend bool, preparedJobs Jobs) {
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

func (r *Renderer) fallback(err error, jobs Jobs) (response *Response) {
	response = new(Response)

	response.Err = err
	response.Results = make(map[string]*JobResult, 0)

	for name, job := range jobs {
		jobResult := new(JobResult)
		id := uuid.NewV4()
		jobResult.HTML = r.getFallbackHTML(job.Name, job.Data, id)
		jobResult.Meta = map[string]string{
			"uuid": id.String(),
		}
		jobResult.OriginalJob = job
		response.Results[name] = jobResult
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

func (r *Renderer) makeRequest(jobs Jobs) (*Response, error) {
	for _, plugin := range r.plugins {
		plugin.WillSendRequest(jobs)
	}

	results, err := r.doRequest(jobs)
	if err != nil {
		return nil, err
	}
	return r.finalize(results), nil
}

func (r *Renderer) doRequest(jobs Jobs) (results JobResults, err error) {
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
		Result map[string]struct {
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

	results = map[string]*JobResult{}
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

func (r *Renderer) finalize(results JobResults) *Response {
	for name, result := range results {
		if result.Err != nil {
			for _, plugin := range r.plugins {
				plugin.OnError(result.Err, Jobs{name: result.OriginalJob})
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
