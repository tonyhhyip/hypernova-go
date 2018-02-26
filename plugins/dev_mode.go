package plugins

import (
	"fmt"

	"github.com/tonyhhyip/hypernova-go"
)

var _ hypernova.Plugin = &DevModePlugin{}

type DevModePlugin struct {
	PluginBase
}

func (p *DevModePlugin) AfterResponse(results []*hypernova.JobResult) []*hypernova.JobResult {
	return results
}

func (p *DevModePlugin) wrapErrors(result *hypernova.JobResult) *hypernova.JobResult {
	if result.Err == nil {
		return result
	}

	result.HTML = fmt.Sprintf(`
		<div style="background-color: #ff5a5f; color: #fff; padding: 12px;">
			<p style="margin: 0">
				<strong>Development Warning!</strong>
			</p>
			<p>
				The <code>%s</code> component failed to render with Hypernova. Error message: %s
			</p>
		</div>
		%s
	`,
		result.OriginalJob.Name,
		result.Err.Error(),
		result.HTML,
	)

	return result
}
