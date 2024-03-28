// Package exec provides the xk6 Modules implementation for running local commands using Javascript
package exec

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"go.k6.io/k6/js/modules"
)

// xk6 build --with xk6-exec=. --with github.com/LeonAdato/xk6-output-statsd
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

func (*EXEC) EnterCommandOnOutput(name string, args []string, outputSearchArray []string, input string) []string {
	// enterCommandOnOutput
	// name, args.
	// outputSearchArray
	var outputArray []string

	fmt.Print(outputSearchArray)

	cmd := exec.Command(name, args...)
	fmt.Println("Starting Initial Command")
	fmt.Println(cmd)

	fmt.Println("Starting Standard In Pipe")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error getting stdin:", err)
		outputArray = append(outputArray, string(err.Error()))
		return outputArray
	}
	fmt.Println(stdin)

	fmt.Println("Starting Standard Out Pipe")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error getting stdout:", err)
		outputArray = append(outputArray, string(err.Error()))
		return outputArray
	}
	fmt.Println(stdout)

	fmt.Println("Starting Start Command")
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting the command:", err)
		outputArray = append(outputArray, string(err.Error()))
		return outputArray
	}

	// output to search for array
	// thing to input
	fmt.Println("Starting Scanner")
	fmt.Println("Output Search Array:", outputSearchArray)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line) // Print the output received from the process

		for _, search := range outputSearchArray {
			if strings.Contains(line, search) {
				fmt.Println("Found:", search)
				goto endScan // Use label to break out of the outer loop
			}
		}
	}

endScan:

	fmt.Println("Starting Write")
	if _, err := stdin.Write([]byte(input + "\n")); err != nil {
		fmt.Println("Error writing to stdin:", err)
		outputArray = append(outputArray, string(err.Error()))
		return outputArray
	}

	stdin.Close()
	fmt.Println("Closed")

	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println(line) // Print the output received from the process
		outputArray = append(outputArray, line)

	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("Command finished with error (This does not always mean a failure):", err)
	}

	return outputArray

}

// Command is a wrapper for Go exec.Command
func (*EXEC) Command(name string, args []string, option CommandOptions) string {
	cmd := exec.Command(name, args...)
	if option.Dir != "" {
		cmd.Dir = option.Dir
	}

	out, err := cmd.CombinedOutput()
	//fmt.Print("Process Output: ")
	//fmt.Print(string(out))

	if err != nil {
		fmt.Print("ERROR: ")
		fmt.Printf(string(err.Error()) + " on command: " + name + " " + strings.Join(args, " "))
	}

	return string(out)
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

	fmt.Print("Process Output: ")
	fmt.Print(string(cmd2Result))

	if err != nil {
		fmt.Print("ERROR: ")
		fmt.Printf(string(err.Error()) + " on command: " + name2 + " " + strings.Join(args2, " "))
	}

	// Wait for both commands to finish
	_ = cmd1.Wait()
	_ = cmd2.Wait()

	// Print the final result
	fmt.Printf("Result:\n%s\n", cmd2Result)

	return string(cmd2Result)
}
