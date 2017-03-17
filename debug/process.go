package debug

import (
	"fmt"
	"os"

	m "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/shirou/gopsutil/process"
)

const (
	localhost = "localhost"
)

type processInfo struct {
	Port string
	Cwd  string
	Pid  int
}

func getPInfos() []processInfo {
	pids, err := process.Pids()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	processes, errors := getProcesses(pids)
	for _, e := range errors {
		logger.Errorf(e)
	}
	return getProcessInfo(processes)
}

func getProcessInfo(processes []*process.Process) []processInfo {
	var infos []processInfo
	for _, p := range processes {
		port, err := getPort(p)
		if err != nil {
			logger.Errorf(err.Error())
			continue
		}
		cwd, err := p.Cwd()
		if err != nil {
			api, err := newAPI(localhost, port)
			if err != nil {
				cwd = "N/A"
			} else {
				msg, err := api.getResponse(&m.APIMessage{MessageType: m.APIMessage_GetProjectRootRequest, ProjectRootRequest: &m.GetProjectRootRequest{}})
				if err != nil {
					cwd = "N/A"
				} else {
					cwd = msg.ProjectRootResponse.ProjectRoot
				}
				api.close()
			}
		}
		infos = append(infos, processInfo{Port: port, Cwd: cwd, Pid: int(p.Pid)})
	}
	return infos
}

func getPort(p *process.Process) (string, error) {
	args, err := p.CmdlineSlice()
	if err == nil {
		if len(args) > 3 {
			return args[len(args)-1], nil
		}
	}
	conns, err := p.Connections()
	if err != nil {
		return "", fmt.Errorf("Cannot find a port for the process %d", p.Pid)
	}
	var port uint32 = 0
	for _, c := range conns {
		p := fmt.Sprintf("%d", c.Laddr.Port)
		api, err := newAPI(localhost, p)
		if err == nil {
			port = c.Laddr.Port
			api.close()
			break
		}
	}
	if port == 0 {
		return "", fmt.Errorf("Cannot find a port for the process %d", p.Pid)
	}
	return fmt.Sprintf("%d", port), nil
}

func getProcesses(pids []int32) ([]*process.Process, []string) {
	var errors []string
	var processes []*process.Process
	for _, pid := range pids {
		if int(pid) == os.Getpid() {
			continue
		}
		p, err := process.NewProcess(pid)
		if err != nil {
			logger.Errorf(err.Error())
		}
		name, err := p.Name()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Process name error for pid: %v. Error: %v", pid, err))
		}
		if name == "gauge" {
			_, err = p.CmdlineSlice()
			if err != nil {
				errors = append(errors, fmt.Sprintf("Process Cmd line slice error for pid: %v. Error: %v", p.Pid, err))
			} else {
				processes = append(processes, p)
			}
		}
	}
	return processes, errors
}
