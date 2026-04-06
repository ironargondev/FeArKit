package config

import (
	"FeArKit/utils"
	"bytes"
	"flag"
	"fmt"
	"github.com/kataras/golog"
	"os"
)

type config struct {
	BackendListen    string     `json:"backendlisten"`
	ClientListen    string      `json:"clientlisten"`
	Salt      string            `json:"salt"`
	Auth      map[string]string `json:"auth"`
	Log       *log              `json:"log"`
	SaltBytes []byte            `json:"-"`
	BackendTLSCert string            	`json:"backend_tls_cert"`
	BackendTLSKey string            	`json:"backend_tls_key"`
	ClientTLSCert string            	`json:"client_tls_cert"`
	ClientTLSKey string            	`json:"client_tls_key"`
	NoGenerate bool             	`json:"noGenerate"`
	AuditLog   *auditLog        	`json:"audit_log"`
}
type log struct {
	Level string `json:"level"`
	Path  string `json:"path"`
	Days  uint   `json:"days"`
}

type auditLog struct {
	Path string `json:"path"`
	Days uint   `json:"days"`
}

// Commit is hash of this commit, for auto upgrade.
var Commit = ``
var Config config
var BuiltPath = `./client/client_%v_%v`

// osFilename maps runtime.GOOS values (sent by clients and the UI) to the OS
// token used in the prebuilt binary filenames.
var osFilename = map[string]string{
	`windows`: `win32nt`,
}

// BuiltClientPath returns the filesystem path for a prebuilt client template.
// It translates OS names and appends the correct extension.
func BuiltClientPath(os, arch string) string {
	token := os
	if mapped, ok := osFilename[os]; ok {
		token = mapped
	}
	p := fmt.Sprintf(BuiltPath, token, arch)
	switch os {
	case `windows`:
		p += `.exe`
	default:
		p += `.elf`
	}
	return p
}
var DevDir string // set via -dev flag; if non-empty, serve frontend from filesystem

func init() {
	golog.SetTimeFormat(`2006/01/02 15:04:05`)

	var (
		err                      error
		configData               []byte
		configPath, salt  		 string
		backendlisten			 string
		clientlisten  			 string
		username, password       string
		logLevel, logPath        string
		logDays                  uint
		noGenerate               bool
	)
	flag.StringVar(&configPath, `config`, `config.json`, `config file path, default: config.json`)
	flag.StringVar(&backendlisten, `backendlisten`, `:9191`, `required, backend listen address, default: 9191`)
	flag.StringVar(&clientlisten, `clientlisten`, `:8000`, `required, listen address, default: :8000`)
	flag.StringVar(&salt, `salt`, ``, `required, salt of server`)
	flag.StringVar(&username, `username`, ``, `username of web interface`)
	flag.StringVar(&password, `password`, ``, `password of web interface`)
	flag.StringVar(&logLevel, `log-level`, `info`, `log level, default: info`)
	flag.StringVar(&logPath, `log-path`, `./logs`, `log file path, default: ./logs`)
	flag.UintVar(&logDays, `log-days`, 7, `max days of logs, default: 7`)
	flag.StringVar(&DevDir, `dev`, ``, `serve frontend from this directory instead of embedded (dev mode)`)
	flag.BoolVar(&noGenerate, `noGenerate`, false, `hide the Generate Client button in the web UI`)
	flag.Parse()

	if len(configPath) > 0 {
		configData, err = os.ReadFile(configPath)
		if err != nil {
			configData, err = os.ReadFile(`config.json`)
			if err != nil {
				fatal(map[string]any{
					`event`:  `CONFIG_LOAD`,
					`status`: `fail`,
					`msg`:    err.Error(),
				})
				return
			}
		}
		err = utils.JSON.Unmarshal(configData, &Config)
		if err != nil {
			fatal(map[string]any{
				`event`:  `CONFIG_PARSE`,
				`status`: `fail`,
				`msg`:    err.Error(),
			})
			return
		}
		if Config.Log == nil {
			Config.Log = &log{
				Level: `info`,
				Path:  `./logs`,
				Days:  7,
			}
		}
		if Config.ClientListen == `` {
			Config.ClientListen = clientlisten
		}
		if Config.BackendListen == `` {
			Config.BackendListen = backendlisten
		}
		if noGenerate {
			Config.NoGenerate = true
		}
	} else {
		Config = config{
			BackendListen: backendlisten,
			ClientListen:  clientlisten,
			Salt:          salt,
			Auth:          map[string]string{username: password},
			NoGenerate:    noGenerate,
			Log: &log{
				Level: logLevel,
				Path:  logPath,
				Days:  logDays,
			},
		}
	}

	if len(Config.Salt) > 24 {
		fatal(map[string]any{
			`event`:  `CONFIG_PARSE`,
			`status`: `fail`,
			`msg`:    `length of salt should less than 24`,
		})
		return
	}
	Config.SaltBytes = []byte(Config.Salt)
	Config.SaltBytes = append(Config.SaltBytes, bytes.Repeat([]byte{25}, 24)...)
	Config.SaltBytes = Config.SaltBytes[:24]

	golog.SetLevel(utils.If(len(Config.Log.Level) == 0, `info`, Config.Log.Level))
}

func fatal(args map[string]any) {
	output, _ := utils.JSON.MarshalToString(args)
	golog.Fatal(output)
}
