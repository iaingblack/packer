package docker

import (
	"fmt"
	"runtime"
	"strings"
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

// 	Current Behaviour
//  ==> docker: Stderr: docker: Error response from daemon: invalid bind mount spec "C:\\Users\\azureuser\\AppData\\Roaming\\packer.d\\tmp\\packer-docker846652819:/packer-files":
// 	Amended Behaviour - "-v c:/xyx/packer.tmp:/packer"
// -v C:/temp/:C:/temp/
//  docker run -v c:/temp/:c:/packer-files/ -d -i -t microsoft/windowsservercore cmd /c
//  docker run -d -i -t microsoft/windowsservercore cmd /c
// Seem to get this error running it sadly...
// Stderr: docker: Error response from daemon: hcsshim::ActivateLayer failed in Win32: The file or directory is corrupted and unreadable. (0x570) 
// docker: Run command: docker run -v C:/Users/azureuser/AppData/Roaming/packer.d/tmp/packer-docker453347183/:c:/packer-files/ -d -i -t microsoft/windowsservercore cmd /c
	if runtime.GOOS == "windows" {
		ui.Say("___________________________+_______________________________")
		tempDir = strings.Replace(tempDir, "\\", "/", -1)
		tempDir = tempDir + "/"
		ui.Say("_DETECTED WINDOWS, REVERSING SLASHES IN PATH TO THIS______")	
		ui.Say("_"+tempDir)
		ui.Say("__________________________________________________________")
	}
	runConfig := ContainerConfig{
		Image:      config.Image,
		RunCommand: config.RunCommand,
		Volumes:    make(map[string]string),
		Privileged: config.Privileged,
	}
	//ui.Say(runConfig.Volumes)
	for host, container := range config.Volumes {
		runConfig.Volumes[host] = container
		ui.Say(" - - - - - - - - - - -  I NEVER GET HERE - - - - - - -")
	}
	if runtime.GOOS == "windows" {
		runConfig.Volumes[tempDir] = "c:/packer-files/"
	} else {
		runConfig.Volumes[tempDir] = "/packer-files"
	}

	
	ui.Say("Starting docker container...")
	containerId, err := driver.StartContainer(&runConfig)
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
