package docker

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepRun struct {
	containerId string
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	tempDir := state.Get("temp_dir").(string)
	ui := state.Get("ui").(packer.Ui)

	runConfig := ContainerConfig{
		Image:      config.Image,
		RunCommand: config.RunCommand,
		Volumes:    make(map[string]string),
		Privileged: config.Privileged,
	}

	for host, container := range config.Volumes {
		runConfig.Volumes[host] = container
	}
	runConfig.Volumes[tempDir] = "/packer-files"

	ui.Say("Starting docker container...")
	containerId, err := driver.StartContainer(&runConfig)
//	This is a fix of sorts I believe. Need to amend the object above but need to research how
// 	package main
// 	import "fmt"
// 	import "runtime"
// 	import "strings"
// 	func main() {
// 	    //Current Behaviour - "-v c:\xyx\packer.tmp:/packer"
// 	    //Amended Behaviour - "-v c:/xyx/packer.tmp:/packer"
// 	    testpath := "C:\\Users\\azureuser\\AppData\\Roaming\\packer.d\\tmp\\packer-docker807709699"
// 	    if runtime.GOOS == "windows" {
// 		fmt.Println(testpath)
// 		result := strings.Replace(testpath, "\\", "/", -1)
// 		fmt.Println(result)
// 	    }
// 	}

	
	if err != nil {
		err := fmt.Errorf("Error running container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Save the container ID
	s.containerId = containerId
	state.Put("container_id", s.containerId)
	ui.Message(fmt.Sprintf("Container ID: %s", s.containerId))
	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.containerId == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// Kill the container. We don't handle errors because errors usually
	// just mean that the container doesn't exist anymore, which isn't a
	// big deal.
	ui.Say(fmt.Sprintf("Killing the container: %s", s.containerId))
	driver.StopContainer(s.containerId)

	// Reset the container ID so that we're idempotent
	s.containerId = ""
}
