package common

import (
	"FeArKit/server/config"
	"FeArKit/utils"
	"FeArKit/utils/melody"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ─── ECS field types ─────────────────────────────────────────────────────────

type ecsEntry struct {
	Timestamp string         `json:"@timestamp"`
	Log       ecsLogLevel    `json:"log"`
	Message   string         `json:"message"`
	Event     ecsEvent       `json:"event"`
	Source    *ecsEndpoint   `json:"source,omitempty"`
	Host      *ecsHost       `json:"host,omitempty"`
	Agent     *ecsAgent      `json:"agent,omitempty"`
	User      *ecsUser       `json:"user,omitempty"`
	Process   *ecsProcess    `json:"process,omitempty"`
	Error     *ecsError      `json:"error,omitempty"`
	Labels    map[string]any `json:"labels,omitempty"`
	FearKit   map[string]any `json:"fearkit,omitempty"`
}

type ecsLogLevel struct {
	Level string `json:"level"`
}

type ecsEvent struct {
	Action   string   `json:"action"`
	Category []string `json:"category,omitempty"`
	Type     []string `json:"type,omitempty"`
	Outcome  string   `json:"outcome,omitempty"`
	Dataset  string   `json:"dataset"`
	Reason   string   `json:"reason,omitempty"`
}

type ecsEndpoint struct {
	IP string `json:"ip,omitempty"`
}

type ecsHost struct {
	Hostname     string   `json:"hostname,omitempty"`
	IP           []string `json:"ip,omitempty"`
	MAC          []string `json:"mac,omitempty"`
	Architecture string   `json:"architecture,omitempty"`
	OS           *ecsOS   `json:"os,omitempty"`
}

type ecsOS struct {
	Type   string `json:"type,omitempty"`
	Family string `json:"family,omitempty"`
}

type ecsAgent struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Version string `json:"version,omitempty"`
}

type ecsUser struct {
	Name string `json:"name,omitempty"`
}

type ecsProcess struct {
	PID int `json:"pid,omitempty"`
}

type ecsError struct {
	Message string `json:"message,omitempty"`
}

// ─── writers ─────────────────────────────────────────────────────────────────

var (
	logWriter   *os.File
	auditWriter *os.File
	auditMu     sync.Mutex
	disposed    bool
)

func init() {
	rotate := func() {
		if disposed {
			golog.SetOutput(os.Stdout)
			return
		}
		now := utils.Now.Add(time.Minute)
		dateStr := now.Format(`2006-01-02`)

		// ── main log ──
		os.Mkdir(config.Config.Log.Path, 0666)
		if logWriter != nil {
			logWriter.Close()
		}
		if config.Config.Log.Level != `disable` {
			var err error
			logWriter, err = os.OpenFile(
				fmt.Sprintf(`%s/%s.log`, config.Config.Log.Path, dateStr),
				os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666,
			)
			if err != nil {
				golog.Warn(buildSimpleJSON(`LOG_INIT`, `fail`, err.Error()))
			} else {
				golog.SetOutput(io.MultiWriter(os.Stdout, logWriter))
			}
			stale := time.Unix(now.Unix()-int64(config.Config.Log.Days*86400), 0)
			os.Remove(fmt.Sprintf(`%s/%s.log`, config.Config.Log.Path, stale.Format(`2006-01-02`)))
		} else {
			golog.SetOutput(os.Stdout)
		}

		// ── audit log ──
		auditPath := config.Config.Log.Path
		auditDays := config.Config.Log.Days
		if config.Config.AuditLog != nil {
			if config.Config.AuditLog.Path != `` {
				auditPath = config.Config.AuditLog.Path
			}
			if config.Config.AuditLog.Days > 0 {
				auditDays = config.Config.AuditLog.Days
			}
		}
		os.Mkdir(auditPath, 0666)
		auditMu.Lock()
		if auditWriter != nil {
			auditWriter.Close()
		}
		var aerr error
		auditWriter, aerr = os.OpenFile(
			fmt.Sprintf(`%s/%s-audit.json`, auditPath, dateStr),
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666,
		)
		auditMu.Unlock()
		if aerr != nil {
			golog.Warn(buildSimpleJSON(`AUDIT_LOG_INIT`, `fail`, aerr.Error()))
		}
		stale := time.Unix(now.Unix()-int64(auditDays*86400), 0)
		os.Remove(fmt.Sprintf(`%s/%s-audit.json`, auditPath, stale.Format(`2006-01-02`)))
	}

	rotate()
	go func() {
		waitSecs := 86400 - (utils.Now.Hour()*3600 + utils.Now.Minute()*60 + utils.Now.Second())
		if waitSecs > 0 {
			<-time.After(time.Duration(waitSecs) * time.Second)
		}
		rotate()
		for range time.NewTicker(time.Second * 86400).C {
			rotate()
		}
	}()
}

// buildSimpleJSON is a minimal JSON builder for bootstrap errors before the
// full logger is available.
func buildSimpleJSON(event, status, msg string) string {
	s, _ := utils.JSON.MarshalToString(map[string]any{
		`event`:  event,
		`status`: status,
		`msg`:    msg,
	})
	return s
}

// ─── legacy text-format log (unchanged shape, extra device fields) ────────────

func getLog(ctx any, event, status, msg string, args map[string]any) string {
	if args == nil {
		args = map[string]any{}
	}
	args[`event`] = event
	if len(msg) > 0 {
		args[`msg`] = msg
	}
	if len(status) > 0 {
		args[`status`] = status
	}
	if ctx != nil {
		var connUUID string
		var targetInfo bool
		switch ctx.(type) {
		case *gin.Context:
			c := ctx.(*gin.Context)
			args[`from`] = GetRealIP(c)
			connUUID, targetInfo = c.Request.Context().Value(`ConnUUID`).(string)
		case *melody.Session:
			s := ctx.(*melody.Session)
			args[`from`] = GetAddrIP(s.GetWSConn().UnderlyingConn().RemoteAddr())
			if deviceConn, ok := args[`deviceConn`]; ok {
				delete(args, `deviceConn`)
				connUUID = deviceConn.(*melody.Session).UUID
				targetInfo = true
			}
		}
		if targetInfo {
			device, ok := Devices.Get(connUUID)
			if ok {
				args[`target`] = map[string]any{
					`name`: device.Hostname,
					`ip`:   device.WAN,
					`os`:   device.OS,
					`arch`: device.Arch,
					`user`: device.Username,
					`pid`:  device.PID,
					`id`:   device.ID,
				}
			}
		}
	}
	output, _ := utils.JSON.MarshalToString(args)
	return output
}

// ─── ECS helpers ─────────────────────────────────────────────────────────────

func ecsCategoryType(action string) ([]string, []string) {
	switch action {
	case `CLIENT_ONLINE`, `SERVER_INIT`, `LISTENER_INIT`:
		return []string{`network`, `session`}, []string{`start`, `connection`}
	case `CLIENT_OFFLINE`:
		return []string{`network`, `session`}, []string{`end`, `connection`}
	case `LOGIN_ATTEMPT`:
		return []string{`authentication`}, []string{`start`}
	case `TERMINAL_CONN`, `TERMINAL_INIT`:
		return []string{`session`}, []string{`start`}
	case `TERMINAL_QUIT`, `TERMINAL_CLOSE`, `TERMINAL_KILL`:
		return []string{`session`}, []string{`end`}
	case `TERMINAL_INPUT`:
		return []string{`session`}, []string{`info`}
	case `DESKTOP_CONN`, `DESKTOP_INIT`:
		return []string{`session`}, []string{`start`}
	case `DESKTOP_QUIT`, `DESKTOP_CLOSE`, `DESKTOP_KILL`:
		return []string{`session`}, []string{`end`}
	case `SCREENSHOT`:
		return []string{`file`}, []string{`access`}
	case `EXEC_COMMAND`, `DOWNLOAD_EXEC`:
		return []string{`process`}, []string{`start`}
	case `PROCESS_KILL`:
		return []string{`process`}, []string{`end`}
	case `PROCESS_INJECT`, `SHELLCODE_EXEC`, `LOAD_ELF`:
		return []string{`process`}, []string{`change`}
	case `REMOVE_FILES`:
		return []string{`file`}, []string{`deletion`}
	case `READ_FILES`, `READ_TEXT_FILE`:
		return []string{`file`}, []string{`access`}
	case `UPLOAD_FILE`:
		return []string{`file`}, []string{`change`}
	case `CALL_DEVICE`:
		return []string{`host`}, []string{`change`}
	case `AGENT_HEARTBEAT`:
		return []string{`network`}, []string{`info`}
	case `CLIENT_UPDATE`:
		return []string{`process`}, []string{`change`}
	default:
		return nil, nil
	}
}

func ecsOutcome(status string) string {
	switch status {
	case `success`:
		return `success`
	case `fail`, `error`:
		return `failure`
	default:
		return `unknown`
	}
}

func ecsMessage(action, status, msg string) string {
	// Named overrides for cleaner messages.
	var base string
	switch action {
	case `CLIENT_ONLINE`:
		base = `agent connected`
	case `CLIENT_OFFLINE`:
		base = `agent disconnected`
	case `AGENT_HEARTBEAT`:
		base = `agent heartbeat`
	case `LOGIN_ATTEMPT`:
		if status == `success` {
			base = `operator login succeeded`
		} else {
			base = `operator login failed`
		}
	default:
		base = strings.ToLower(strings.ReplaceAll(action, `_`, ` `))
		switch status {
		case `fail`:
			base += ` failed`
		case `error`:
			base += ` error`
		case `success`:
			base += ` succeeded`
		}
	}
	if msg != `` {
		base += `: ` + msg
	}
	return base
}

// buildECS constructs a fully-populated ECS log entry from a call context.
// It must be called BEFORE getLog since getLog mutates the args map in place
// (it deletes "deviceConn" and adds "event"/"msg"/"status"/"from"/"target").
func buildECS(ctx any, level, action, status, msg string, args map[string]any) *ecsEntry {
	cats, types := ecsCategoryType(action)
	reason := msg

	entry := &ecsEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Log:       ecsLogLevel{Level: level},
		Message:   ecsMessage(action, status, msg),
		Event: ecsEvent{
			Action:   strings.ToLower(action),
			Category: cats,
			Type:     types,
			Outcome:  ecsOutcome(status),
			Dataset:  `fearkit.audit`,
			Reason:   reason,
		},
	}

	fearkit := map[string]any{}
	labels := map[string]any{}

	// ── source IP and target device from context ──────────────────────────────
	var targetConnUUID string
	switch v := ctx.(type) {
	case *gin.Context:
		entry.Source = &ecsEndpoint{IP: GetRealIP(v)}
		if uuid, ok := v.Request.Context().Value(`ConnUUID`).(string); ok {
			targetConnUUID = uuid
		}
	case *melody.Session:
		entry.Source = &ecsEndpoint{IP: GetAddrIP(v.GetWSConn().UnderlyingConn().RemoteAddr())}
		if dc, ok := args[`deviceConn`]; ok {
			if s, ok := dc.(*melody.Session); ok {
				targetConnUUID = s.UUID
			}
		} else {
			// ctx is the client session itself (e.g. disconnect handler)
			targetConnUUID = v.UUID
		}
	}

	// ── populate host/agent/user from device record ───────────────────────────
	if targetConnUUID != `` {
		fearkit[`session_id`] = targetConnUUID
		if device, ok := Devices.Get(targetConnUUID); ok {
			h := &ecsHost{
				Hostname:     device.Hostname,
				Architecture: device.Arch,
			}
			if device.WAN != `` {
				h.IP = []string{device.WAN}
			}
			if device.LAN != `` && device.LAN != device.WAN {
				h.IP = append(h.IP, device.LAN)
			}
			if device.MAC != `` {
				h.MAC = []string{device.MAC}
			}
			if device.OS != `` {
				h.OS = &ecsOS{Type: device.OS, Family: device.OS}
			}
			entry.Host = h
			entry.Agent = &ecsAgent{
				ID:      targetConnUUID,
				Type:    `fearkit-client`,
				Version: config.Commit,
			}
			if device.Username != `` {
				entry.User = &ecsUser{Name: device.Username}
			}
			if device.PID > 0 {
				entry.Process = &ecsProcess{PID: device.PID}
			}
			fearkit[`device_id`] = device.ID
			fearkit[`lan_ip`] = device.LAN
			fearkit[`uptime_s`] = device.Uptime
		}
	}

	// ── error field for failures ──────────────────────────────────────────────
	if (status == `fail` || status == `error`) && msg != `` {
		entry.Error = &ecsError{Message: msg}
	}

	// ── extra args into labels / fearkit ─────────────────────────────────────
	// Keys added later by getLog are skipped.
	skip := map[string]bool{
		`event`: true, `msg`: true, `status`: true,
		`from`: true, `target`: true, `deviceConn`: true,
	}
	for k, v := range args {
		if skip[k] {
			continue
		}
		switch k {
		case `latency_ms`:
			fearkit[`latency_ms`] = v
		case `pid`:
			fearkit[`target_pid`] = toIntAny(v)
		case `cmd`, `args`, `url`, `path`, `file`, `files`,
			`act`, `os`, `arch`, `commit`, `user`, `error`:
			labels[k] = v
		case `listen`, `protocol`:
			labels[k] = v
		case `client`:
			labels[`client`] = v
		case `server`:
			labels[`server`] = v
		case `device`:
			labels[`device`] = v
		default:
			labels[k] = v
		}
	}

	if len(labels) > 0 {
		entry.Labels = labels
	}
	if len(fearkit) > 0 {
		entry.FearKit = fearkit
	}

	return entry
}

func toIntAny(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

func writeAudit(entry *ecsEntry) {
	if entry == nil {
		return
	}
	data, err := utils.JSON.Marshal(entry)
	if err != nil {
		return
	}
	auditMu.Lock()
	defer auditMu.Unlock()
	if auditWriter != nil {
		auditWriter.Write(data)
		auditWriter.Write([]byte{'\n'})
	}
}

// ─── public logging functions ─────────────────────────────────────────────────

// Info logs at info level and writes an ECS audit record.
func Info(ctx any, event, status, msg string, args map[string]any) {
	// buildECS must run before getLog (getLog mutates args).
	writeAudit(buildECS(ctx, `info`, event, status, msg, args))
	golog.Infof(getLog(ctx, event, status, msg, args))
}

// Warn logs at warn level and writes an ECS audit record.
func Warn(ctx any, event, status, msg string, args map[string]any) {
	writeAudit(buildECS(ctx, `warn`, event, status, msg, args))
	golog.Warnf(getLog(ctx, event, status, msg, args))
}

// Error logs at error level and writes an ECS audit record.
func Error(ctx any, event, status, msg string, args map[string]any) {
	writeAudit(buildECS(ctx, `error`, event, status, msg, args))
	golog.Error(getLog(ctx, event, status, msg, args))
}

// Fatal logs at fatal level, writes an ECS audit record, and exits.
func Fatal(ctx any, event, status, msg string, args map[string]any) {
	writeAudit(buildECS(ctx, `fatal`, event, status, msg, args))
	golog.Fatalf(getLog(ctx, event, status, msg, args))
}

// Debug logs at debug level with extended context. Debug events are NOT
// written to the audit log — they are for operator troubleshooting only.
func Debug(ctx any, event, status, msg string, args map[string]any) {
	golog.Debugf(getLog(ctx, event, status, msg, args))
}

// Audit writes an ECS audit record without emitting to the regular log.
// Use for periodic events (heartbeat) that would clutter the main log
// but belong in the structured audit trail.
func Audit(ctx any, level, event, status, msg string, args map[string]any) {
	writeAudit(buildECS(ctx, level, event, status, msg, args))
}

func CloseLog() {
	disposed = true
	golog.SetOutput(os.Stdout)
	if logWriter != nil {
		logWriter.Close()
		logWriter = nil
	}
	auditMu.Lock()
	defer auditMu.Unlock()
	if auditWriter != nil {
		auditWriter.Close()
		auditWriter = nil
	}
}
