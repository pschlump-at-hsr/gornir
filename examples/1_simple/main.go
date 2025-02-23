// this is the simplest example possible
package main

import (
	"context"
	"os"

	"github.com/pschlump-at-hsr/gornir/pkg/gornir"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/connection"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/inventory"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/logger"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/output"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/runner"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/task"
)

func main() {
	// Instantiate a logger plugin
	log := logger.NewLogrus(false)

	// Load the inventory using the FromYAMLFile plugin
	file := "/go/src/github.com/pschlump-at-hsr/gornir/examples/hosts.yaml"
	plugin := inventory.FromYAML{HostsFile: file}
	inv, err := plugin.Create()
	if err != nil {
		log.Fatal(err)
	}

	rnr := runner.Sorted()

	gr := gornir.New().WithInventory(inv).WithLogger(log).WithRunner(rnr)

	// Open an SSH connection towards the devices
	results, err := gr.RunSync(
		context.Background(),
		&connection.SSHOpen{},
	)
	if err != nil {
		log.Fatal(err)
	}
	output.RenderResults(os.Stdout, results, "Connecting to devices via ssh", true)

	// defer closing the SSH connection we just opened
	defer func() {
		results, err = gr.RunSync(
			context.Background(),
			&connection.SSHClose{},
		)
		if err != nil {
			log.Fatal(err)
		}
		output.RenderResults(os.Stdout, results, "Close ssh connection", true)
	}()

	// Following call is going to execute the task over all the hosts using the runner.Parallel runner.
	// Said runner is going to handle the parallelization for us. Gornir.RunS is also going to block
	// until the runner has completed executing the task over all the hosts
	results, err = gr.RunSync(
		context.Background(),
		&task.RemoteCommand{Command: "ip addr | grep \\/24 | awk '{ print $2 }'"},
	)
	if err != nil {
		log.Fatal(err)
	}
	// next call is going to print the result on screen
	output.RenderResults(os.Stdout, results, "What is my ip?", true)

	// Now we upload a file. This shows how the ssh connection is shared across tasks of same or different type
	results, err = gr.RunSync(
		context.Background(),
		&task.SFTPUpload{Src: "/etc/hosts", Dst: "/tmp/asd"},
	)
	if err != nil {
		log.Fatal(err)
	}
	output.RenderResults(os.Stdout, results, "Upload File", true)
}
