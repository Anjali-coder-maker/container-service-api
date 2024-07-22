package utils

import "os/exec"

type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func ExecuteCommand(command string, args ...string) CommandResponse {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CommandResponse{Output: string(output), Error: err.Error()}
	}
	return CommandResponse{Output: string(output)}
}
