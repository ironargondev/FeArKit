package executable

import (
	"net/http"
	"time"

	"FeArKit/server/handler/utility"
	"FeArKit/server/common"
	"FeArKit/modules"
	"FeArKit/utils/melody"
	"FeArKit/utils"

	"github.com/gin-gonic/gin"
)

// ListDeviceProcesses will list processes on remote client
func DownloadAndExecute(ctx *gin.Context) {
	var form struct {
		Url string `json:"url" yaml:"url" form:"url" binding:"required"`
		TargetPath string `json:"path" yaml:"path" form:"path"`
	}
	target, ok := utility.CheckForm(ctx, &form)
	if !ok {
		return
	}
	if len(form.Url) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, modules.Packet{Code: -1, Msg: `${i18n|COMMON.INVALID_PARAMETER}`})
		return
	}

	trigger := utils.GetStrUUID()
	common.Info(ctx, `executable`, `exec`, ``, map[string]any{
		`url`: form.Url,
		`targetpath`: form.TargetPath,
	})
	common.SendPackByUUID(modules.Packet{Act: `DOWNLOAD_EXEC`, Data: gin.H{`url`: form.Url, `path`: form.TargetPath}, Event: trigger}, target)
	ok = common.AddEventOnce(func(p modules.Packet, _ *melody.Session) {
		if p.Code != 0 {
			common.Warn(ctx, `DOWNLOAD_EXEC`, `fail`, p.Msg, map[string]any{
				`url`: form.Url,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, modules.Packet{Code: 1, Msg: p.Msg})
		} else {
			common.Info(ctx, `DOWNLOAD_EXEC`, `success`, ``, map[string]any{
				`url`: form.Url,
			})
			ctx.JSON(http.StatusOK, modules.Packet{Code: 0})
		}
	}, target, trigger, 5*time.Second)
	if !ok {
		common.Warn(ctx, `DOWNLOAD_EXEC`, `fail`, `timeout`, map[string]any{
			`url`: form.Url,
		})
		ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, modules.Packet{Code: 1, Msg: `${i18n|COMMON.RESPONSE_TIMEOUT}`})
	}
}

