# Hypernova Go

Go Client for [hypernova](https://github.com/airbnb/hypernova)

## Install

[Glide](https://glide.sh) is recommend to use.
```bash
glide get github.com/tonyhhyip/hypernova-go
```

## Usage

```go

package main
import (
	"fmt"
	"github.com/tonyhhyip/hypernova-go"
)

func main() {
	renderer := hypernova.NewRenderer("http://localhost:3000/batch")
	job := &hypernova.Job {
		Name: "moduleName",
		Data: map[string]interface{} {
			"foo": "bar",
		},
		Metadata: map[string]string{},
	}
	resp := renderer.AddJob("viewId", job)
	fmt.Println(resp.Results["viewId"].HTML)
}

```
