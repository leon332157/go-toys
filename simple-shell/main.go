package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Process struct {
	Cmd      string
	Args     []string
	pipe_in  bool
	pipe_out bool
	pipe_r   *os.File
	pipe_w   *os.File
	execCmd  *exec.Cmd
}

type ProcessList []Process

func writePrompt() {
	os.Stdout.Write([]byte("$ "))
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	is_quit := false
	for {
		if is_quit {
			break
		}
		writePrompt()
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		line = strings.TrimSuffix(line, "\n")
		processList := parseInput(line)
		is_quit = runCommands(processList)
	}
}

func parseInput(line string) ProcessList {
	var processList ProcessList
	if line == "" { // if this line is empty, return empty process list
		return processList
	}
	cmd_indep := strings.Split(line, ";") // independent command chains
	for _, indep := range cmd_indep {
		cmd_pipes := strings.Split(indep, "|") // dependent command chains
		for idx, dep_cmds := range cmd_pipes {
			currProcess := Process{}
			args := strings.Fields(dep_cmds)
			if len(args) == 0 {
				break
			}
			currProcess.Cmd = args[0]
			if len(args) > 1 {
				currProcess.Args = args[1:]
			}

			currProcess.pipe_out = true // pipe out by default
			if idx == len(cmd_pipes) - 1 { // if it's the last command, do not pipe out
				currProcess.pipe_out = false
			}

			currProcess.pipe_in = false
			if idx > 0 { // if second onwards, pipe in
				currProcess.pipe_in = true
			}

			processList = append(processList, currProcess)
		}
	}
	return processList
}

func runCommands(command_list ProcessList) bool {
	is_quit := false
	var prev *Process = nil
	var waitList ProcessList;
	for _, currCmd := range command_list {
		currExec := exec.Command(currCmd.Cmd, currCmd.Args...)

		if currCmd.Cmd == "quit" { // if any command is quit, break execution
			is_quit = true
			break
		}

		currExec.Stderr = os.Stderr

		if currCmd.pipe_out {
			pr, pw, err := os.Pipe()
			if err != nil {
				panic("creating pipe failed")
			}

			currCmd.pipe_r = pr
			currCmd.pipe_w = pw
			currExec.Stdout = pw
		} else {
			currExec.Stdout = os.Stdout
		}

		if currCmd.pipe_in && prev != nil {
			currExec.Stdin = prev.pipe_r
		} else {
			currExec.Stdin = os.Stdin
		}
		
		err := currExec.Start()
		if err != nil {
			fmt.Printf("starting process %v failed %v", currExec.Args, err)
		}
		currCmd.execCmd = currExec
		prev = &currCmd
		waitList = append(waitList, currCmd)
	}

	for _, cmd := range waitList {
		err := cmd.execCmd.Wait()
		if err != nil {
			fmt.Printf("command %v wait failed %v\n", cmd.Args, err)
		}
		// close the write end and read end
		cmd.pipe_w.Close() 
		cmd.pipe_r.Close()
	}

	return is_quit
}
