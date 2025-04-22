
package main

import (
	"flag"
	"fmt"
	"os"
	"bytes"
	"encoding/hex"
	"encoding/json"

	"FeArKit/utils"
	"crypto/aes"
	"crypto/cipher"
	"math/big"
	"errors"
	"strings"
)

type clientCfg struct {
	Secure bool   `json:"secure"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Path   string `json:"path"`
	UUID   string `json:"uuid"`
	Key    string `json:"key"`
}

var (
	ErrTooLargeEntity = errors.New(`length of data can not excess buffer size`)
)

func main() {
	var (
		hostFlag, pathFlag, inFile, outFile, saltString, configPath string
		portFlag                                                    uint
		secureFlag, stdOut, debug                                   bool
	)
	flag.StringVar(&hostFlag, "host", "", "server host (required)")
	flag.UintVar(&portFlag, "port", 0, "server port (required)")
	flag.StringVar(&pathFlag, "path", "/", "server path")
	flag.BoolVar(&secureFlag, "secure", false, "enable secure connection")
	flag.StringVar(&inFile, "in", "", "path to the input file")
	flag.StringVar(&outFile, "out", "", "path to the output file")
	flag.BoolVar(&stdOut, "stdout", false, "Print hex encoded config")
	flag.BoolVar(&debug, "debug", false, "Print debug information")
	flag.StringVar(&saltString, "salt", "", "salt of server")
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.Parse()

	// Debug output to help troubleshoot
	if debug {
		fmt.Println("Command line arguments:", os.Args)
		fmt.Printf("Parsed flags: host=%s, port=%d, path=%s, secure=%v, in=%s, out=%s, salt=%s, configpath=%s\n",
			hostFlag, portFlag, pathFlag, secureFlag, inFile, outFile, saltString, configPath)
	}

	if hostFlag == "" || portFlag == 0 {
		flag.Usage()
		os.Exit(1)
	}
	var err error
	if len(configPath) > 0 {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("failed to read config file: %v", err)
			os.Exit(1)
		}
		var cfg map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(configData))
		if err := decoder.Decode(&cfg); err != nil {
			fmt.Printf("failed to unmarshal config file: %v", err)
			os.Exit(1)
		}
		salt, ok := cfg["salt"].(string)
		if !ok {
			fmt.Printf("failed to get salt from config file")
			os.Exit(1)
		}
		saltString = salt
	} else {
		if saltString == "" {
			fmt.Printf("salt is required")
			os.Exit(1)
		}
	}

	saltBytes := []byte(saltString)
	saltBytes = append(saltBytes, bytes.Repeat([]byte{25}, 24)...)
	saltBytes = saltBytes[:24]

	var input, output *os.File
	if !stdOut {
		// Open the input file.
		var err error
		input, err = os.Open(inFile)
		if err != nil {
			fmt.Printf("failed to open input file: %v", err)
			os.Exit(1)
		}
		defer input.Close()

		// Create the output file.
		output, err = os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Printf("failed to create output file: %v", err)
			os.Exit(1)
		}
		defer output.Close()
	}
	// Generate configuration data.
	clientUUID := utils.GetUUID()
	clientKey, err := encAES(clientUUID, []byte(saltBytes))
	if err != nil {
		fmt.Printf("failed to generate client key: %v", err)
		os.Exit(1)
	}
	clientConfigJson := clientCfg{
		Secure: secureFlag,
		Host:   hostFlag,
		Port:   int(portFlag),
		Path:   pathFlag,
		UUID:   hex.EncodeToString(clientUUID),
		Key:    hex.EncodeToString(clientKey),
	}
	if debug {
		fmt.Printf("Config before encrypting: %+v\n", clientConfigJson)
	}
	cfgBytes, err := genConfig(clientConfigJson)
	if err != nil {
		fmt.Printf("failed to generate config: %v", err)
		os.Exit(1)
	}
	if stdOut {
		// Print config buffer in \x encoded hex
		var hexBuffer strings.Builder
		for i := 0; i < len(cfgBytes); i++ {
			hexBuffer.WriteString(fmt.Sprintf("\\x%02x", cfgBytes[i]))
		}
		fmt.Printf("Config Buffer (hex): %v", hexBuffer.String())

		fmt.Println()
		return
	}
	// Read the input file and replace the placeholder buffer with the generated configuration.
	placeholder := bytes.Repeat([]byte{'\x19'}, 384)
	var prevBuffer []byte
	var templateFound = false
	buf := make([]byte, 1024)
	for {
		n, readErr := input.Read(buf)
		chunk := buf[:n]
		tempBuffer := append(prevBuffer, chunk...)
		if bytes.Index(tempBuffer, placeholder) > -1 {
			fmt.Println("Found placeholder, replacing with config...")
			tempBuffer = bytes.ReplaceAll(tempBuffer, placeholder, cfgBytes)
			templateFound = true
		}
		// Write out complete data from previous iteration.
		if len(tempBuffer) > len(prevBuffer) {
			if _, err := output.Write(tempBuffer[:len(tempBuffer)-len(prevBuffer)]); err != nil {
				fmt.Printf("failed to write to output: %v", err)
				os.Exit(1)
			}
		}
		prevBuffer = tempBuffer[len(tempBuffer)-len(prevBuffer):]
		if readErr != nil {
			break
		}
	}
	// Write any remaining data.
	if len(prevBuffer) > 0 {
		if _, err := output.Write(prevBuffer); err != nil {
			fmt.Printf("failed to write remaining data: %v", err)
			os.Exit(1)
		}
	}
	if !templateFound {
		fmt.Println("No placeholder found in the input file.")
		os.Exit(1)
	}
	fmt.Println("File has been patched successfully.")
}

func genConfig(cfg clientCfg) ([]byte, error) {
	data, err := utils.JSON.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	key := utils.GetUUID()
	data, err = encAES(data, key)
	if err != nil {
		return nil, err
	}
	final := append(key, data...)
	if len(final) > 384-2 {
		return nil, ErrTooLargeEntity
	}

	// Get the length of encrypted buffer as a 2-byte big-endian integer.
	// And append encrypted buffer to the end of the data length.
	dataLen := big.NewInt(int64(len(final))).Bytes()
	dataLen = append(bytes.Repeat([]byte{'\x00'}, 2-len(dataLen)), dataLen...)

	// If the length of encrypted buffer is less than 384,
	// append the remaining bytes with random bytes.
	final = append(dataLen, final...)
	for len(final) < 384 {
		final = append(final, utils.GetUUID()...)
	}
	return final[:384], nil
}

func encAES(data []byte, key []byte) ([]byte, error) {
	hash, _ := utils.GetMD5(data)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, hash)
	encBuffer := make([]byte, len(data))
	stream.XORKeyStream(encBuffer, data)
	return append(hash, encBuffer...), nil
}