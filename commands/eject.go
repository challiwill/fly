package commands

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/concourse/atc"
	"github.com/concourse/fly/commands/internal/flaghelpers"
	"github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"
)

type EjectResourceCommand struct {
	Resource flaghelpers.ResourceFlag `short:"r" long:"resource" required:"true" value-name:"PIPELINE/RESOURCE" description:"Name of a resource to delete version from"`
	Version  *atc.Version             `short:"n" long:"version"  required:"true" value-name:"VERSION" description:"Version of a resource to delete"`
}

func (command *EjectResourceCommand) Execute(args []string) error {
	client, err := rc.TargetClient(Fly.Target)
	if err != nil {
		return err
	}
	err = rc.ValidateClient(client, Fly.Target, false)
	if err != nil {
		return err
	}

	version := *command.Version
	if len(version) > 1 {
		return errors.New("you can only delete one version at a time")
	}

	// this is just to get the version in a way that is printable
	var ver, ref string
	for r, v := range version {
		ver = v
		ref = r
	}

	page := &concourse.Page{Limit: 100}
dance:
	for page != nil {
		versionedResources, pagination, _, err := client.ResourceVersions(command.Resource.PipelineName, command.Resource.ResourceName, *page)
		if err != nil {
			return err
		}

		for _, vr := range versionedResources {
			if reflect.DeepEqual(version, vr.Version) {
				//TODO delete volume from workers
				// - this is hard. Currently there's only access to the baggagecollector when it gets kicked off in atccmd/command.go
				// after that we never have a reference to it. We would need the baggagecollector to expose its own endpoint that we
				// could hit, or let workers expose their volumes. Right now it seems all volume expiration happens through ttl's
				// getting updated because they are no longer the newest version.
				//TODO reap volume from db
				found, err := client.DeleteResourceVersion(command.Resource.PipelineName, command.Resource.ResourceName, vr.ID)
				if err != nil {
					return err
				}
				if !found {
					break dance
				}

				fmt.Printf("deleted '%s' version '%s:%s'\n", command.Resource.ResourceName, ref, ver)
				return nil
			}
		}
		page = pagination.Next
	}

	return fmt.Errorf("pipeline '%s' or resource '%s' or version '%s:%s' not found\n", command.Resource.PipelineName, command.Resource.ResourceName, ref, ver)
}
