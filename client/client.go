package main

import (
	"FeArKit/client/config"
	"FeArKit/client/core"
	"FeArKit/utils"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"
	"fmt"

	"github.com/kataras/golog"
)
const (
	AppName     = "FeArKit"
	AppVersion  = "1.0.0"
)


func init() {
	golog.SetTimeFormat(`2006/01/02 15:04:05`)
	golog.SetLevel("critical")
	golog.Debug("Initializing client...")

	ConfigBuffer := config.ConfigBuffer
	if len(strings.Trim(ConfigBuffer, "\x19")) == 0 {
		golog.Debug("Error: Internal config buffer is empty")
		//Check arguments
		if len(os.Args) > 1 {
			filePath := os.Args[1]
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				golog.Debug("File exists:", filePath)
				contents, err := os.ReadFile(filePath)
				if err != nil {
					golog.Debug("Failed to read file:", err)
					return
				}
				ConfigBuffer = string(contents)
			} else {
				golog.Debug("File does not exist or is a directory:", filePath)
				ConfigBuffer = filePath
			}
		} else {
			golog.Debug("No file path provided in arguments.")
			os.Exit(1)
		}
	}

	// Print config buffer in \x encoded hex
	var hexBuffer strings.Builder
	for i := 0; i < len(config.ConfigBuffer); i++ {
		hexBuffer.WriteString(fmt.Sprintf("\\x%02x", config.ConfigBuffer[i]))
	}
	golog.Debug("Config Buffer (hex):", hexBuffer.String())

	// Convert first 2 bytes to int, which is the length of the encrypted config.
	dataLen := int(big.NewInt(0).SetBytes([]byte(config.ConfigBuffer[:2])).Uint64())
	if dataLen > len(config.ConfigBuffer)-2 {
		golog.Debug("Error: Invalid config length")
		os.Exit(1)
		return
	}
	cfgBytes := utils.StringToBytes(config.ConfigBuffer, 2, 2+dataLen)
	cfgBytes, err := decrypt(cfgBytes[16:], cfgBytes[:16])
	if err != nil {
		golog.Debug("Error: Failed to decrypt config:", err)
		os.Exit(1)
		return
	}

	err = utils.JSON.Unmarshal(cfgBytes, &config.Config)
	if err != nil {
		golog.Debug("Error: Failed to unmarshal config:", err)
		os.Exit(1)
		return
	}

	config.Config.ClientUptime = time.Now().UTC().Unix()

	if strings.HasSuffix(config.Config.Path, `/`) {
		config.Config.Path = config.Config.Path[:len(config.Config.Path)-1]
	}
	configJSON, err := utils.JSON.MarshalIndent(config.Config, "", "  ")
	if err != nil {
		golog.Debug("Failed to marshal config:", err)
	} else {
		golog.Debug("Loaded config:\n", string(configJSON))
	}
	HideConsoleWindow()
}

func main() {
	golog.Debug("Starting client...")
	update()
	core.Start()

}

func update() {
	selfPath, err := os.Executable()
	if err != nil {
		selfPath = os.Args[0]
	}
	if len(os.Args) > 1 && os.Args[1] == `--update` {
		if len(selfPath) <= 4 {
			return
		}
		destPath := selfPath[:len(selfPath)-4]
		thisFile, err := os.ReadFile(selfPath)
		if err != nil {
			return
		}
		os.WriteFile(destPath, thisFile, 0755)
		cmd := exec.Command(destPath, `--clean`)
		if cmd.Start() == nil {
			os.Exit(0)
			return
		}
	}
	if len(os.Args) > 1 && os.Args[1] == `--clean` {
		<-time.After(3 * time.Second)
		os.Remove(selfPath + `.tmp`)
	}
}

func decrypt(data []byte, key []byte) ([]byte, error) {
	// MD5[16 bytes] + Data[n bytes]
	dataLen := len(data)
	if dataLen <= 16 {
		return nil, utils.ErrEntityInvalid
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, data[:16])
	decBuffer := make([]byte, dataLen-16)
	stream.XORKeyStream(decBuffer, data[16:])
	hash, _ := utils.GetMD5(decBuffer)
	if !bytes.Equal(hash, data[:16]) {
		return nil, utils.ErrFailedVerification
	}
	return decBuffer[:dataLen-16], nil
}
