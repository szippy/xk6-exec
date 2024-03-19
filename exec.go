// Package exec provides the xk6 Modules implementation for running local commands using Javascript
package exec

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/exec", new(RootModule))
}

// RootModule is the global module object type. It is instantiated once per test
// run and will be used to create `k6/x/exec` module instances for each VU.
type RootModule struct{}

// EXEC represents an instance of the EXEC module for every VU.
type EXEC struct {
	vu modules.VU
}

// CommandOptions contains the options that can be passed to command.
type CommandOptions struct {
	Dir string
}

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &EXEC{}
)

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &EXEC{vu: vu}
}

// Exports implements the modules.Instance interface and returns the exports
// of the JS module.
func (exec *EXEC) Exports() modules.Exports {
	return modules.Exports{Default: exec}
}

// Command is a wrapper for Go exec.Command
func (*EXEC) Command(name string, args []string, option CommandOptions) string {
	fmt.Print("EXEC COMMAND")
	cmd := exec.Command(name, args...)
	if option.Dir != "" {
		cmd.Dir = option.Dir
	}
	// print output
	out, err := cmd.Output()
	fmt.Print("Process Output: ")
	fmt.Print(string(out))

	if err != nil {
		fmt.Print("ERROR: ")
		fmt.Printf(string(err.Error()) + " on command: " + name + " " + strings.Join(args, " "))
	}
	return string(out)
	// get right pipe.
}

func (*EXEC) PipeCommand(name1 string, args1 []string, name2 string, args2 []string, option CommandOptions) string {
	// ex ls -l | grep "go"
	//name1 = "ls"
	//args1 = []string{"-l"}
	//name2 = "grep"
	//args2 = []string{"go"}
	// set the command before the pipe
	cmd1 := exec.Command(name1, args1...)
	if option.Dir != "" {
		cmd1.Dir = option.Dir
	}

	fmt.Print("Command 1: ")
	fmt.Print(cmd1)

	// set the command after the pipe
	cmd2 := exec.Command(name2, args2...)
	if option.Dir != "" {
		cmd2.Dir = option.Dir
	}

	fmt.Print("Command 2: ")
	fmt.Print(cmd2)

	// Get the output of the command that runs after the pipe and connect it as input to first command
	cmd2.Stdin, _ = cmd1.StdoutPipe()

	// Get the output of the second command
	cmd2Output, _ := cmd2.StdoutPipe()

	// Start both commands
	_ = cmd2.Start()
	_ = cmd1.Start()

	// Read the output of the second command
	cmd2Result, err := io.ReadAll(cmd2Output)
	if err != nil {
		fmt.Printf("Error reading command output: %v\n", err)
		return string(err.Error())
	}

	// Wait for both commands to finish
	_ = cmd1.Wait()
	_ = cmd2.Wait()

	// Print the final result
	fmt.Printf("Result:\n%s\n", cmd2Result)

	return string(cmd2Result)
}
