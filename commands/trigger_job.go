package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/concourse/fly/commands/internal/flaghelpers"
	"github.com/concourse/fly/eventstream"
	"github.com/concourse/fly/rc"
	"github.com/concourse/fly/ui"
)

type TriggerJobCommand struct {
	Job   flaghelpers.JobFlag `short:"j" long:"job" required:"true" value-name:"PIPELINE/JOB" description:"Name of a job to trigger"`
	Watch bool                `short:"w" long:"watch" description:"Start watching the build output"`
}

func (command *TriggerJobCommand) Execute(args []string) error {
	pipelineName, jobName := command.Job.PipelineName, command.Job.JobName

	client, err := rc.TargetClient(Fly.Target)
	if err != nil {
		return err
	}

	err = rc.ValidateClient(client, Fly.Target, false)
	if err != nil {
		return err
	}

	build, err := client.CreateJobBuild(pipelineName, jobName)
	if err != nil {
		return err
	}
	fmt.Printf("started %s/%s #%s\n", pipelineName, jobName, build.Name)

	if command.Watch {
		terminate := make(chan os.Signal, 1)

		go func(terminate <-chan os.Signal) {
			<-terminate
			fmt.Fprintf(os.Stderr, "\ndetached, build is still running...\n")
			fmt.Fprintf(os.Stderr, "re-attach to it with:\n\n")
			fmt.Fprintf(os.Stderr, "    "+ui.Embolden(fmt.Sprintf("fly -t %s watch -j %s/%s -b %s\n\n", Fly.Target, pipelineName, jobName, build.Name)))
			os.Exit(2)
		}(terminate)

		signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

		fmt.Println("")
		eventSource, err := client.BuildEvents(fmt.Sprintf("%d", build.ID))
		if err != nil {
			return err
		}

		exitCode := eventstream.Render(os.Stdout, eventSource)

		eventSource.Close()

		os.Exit(exitCode)
	}

	return nil
}
