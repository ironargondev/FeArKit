<h1 align="center">KnownMalwareProject</h1>

Forked from **Spark** (https://github.com/XZB-1248/Spark), a free, safe, open-source, web-based, cross-platform and full-featured RAT (Remote Administration Tool)
that allow you to control all your devices via browser anywhere.

Modifications by Iron Argon Development

---

## Disclaimer

**THIS PROJECT, ITS SOURCE CODE, AND ITS RELEASES SHOULD ONLY BE USED FOR EDUCATIONAL PURPOSES. THE PURPOSE OF THIS PROJECT IS TO TRAIN THREAT HUNTERS, REVERSE ENGINEERS AND CREATE AV SIGNATURES**
<br />
**YOU SHALL USE THIS PROJECT AT YOUR OWN RISK.**
<br />
**THE AUTHORS AND DEVELOPERS ARE NOT RESPONSIBLE FOR ANY DAMAGE CAUSED BY YOUR MISUSE OF THIS PROJECT.**

---

## 🚀 Quick Start

### Binary Execution

1. Download the executable from the [releases](https://github.com/ironargondev/FeArKit/releases) page.
2. Follow the [Configuration](#configuration) instructions.
3. Run the executable and access the web interface at `http://IP:Port`.
4. Generate a client and run it on the target device.
5. Start managing your devices!

---

## ⚙️ Configuration

The configuration file `config.json` should be in the same directory as the executable.

**Example:**

```json
{
    "backendlisten": ":9191",
    "clientlisten": ":8000",
    "salt": "123456abcdef123456",
    "auth": {
        "username": "password"
    },
    "log": {
        "level": "info",
        "path": "./logs",
        "days": 7
    },
    "noGenerate": false
}
```

### Main Parameters:
- **`backendlisten`** (required): Address for the web UI and API server. Format `IP:Port` (e.g. `:9191`).
- **`clientlisten`** (required): Address for the C2 agent WebSocket listener. Format `IP:Port` (e.g. `:8000`).
- **`salt`** (required): Max length 24 characters. After modification, all clients need to be regenerated.
- **`auth`** (optional): Authentication credentials (`username:password`).
  - Hashed passwords are recommended (`$algorithm$hashed-password`).
  - Supported algorithms: `sha256`, `sha512`, `bcrypt`.
- **`log`** (optional): Logging configuration.
  - `level`: `disable`, `fatal`, `error`, `warn`, `info`, `debug`.
  - `path`: Log directory (default: `./logs`).
  - `days`: Log retention days (default: `7`).
- **`noGenerate`** (optional): Set to `true` to hide the Generate Client button in the web UI.

---

## 🛠️ Features

| Feature/OS        | Windows | Linux | MacOS |
|-------------------|---------|-------|-------|
| Process Manager   | ✔       | ✔     | ✔     |
| Kill Process      | ✔       | ✔     | ✔     |
| Network Traffic   | ✔       | ✔     | ✔     |
| File Explorer     | ✔       | ✔     | ✔     |
| File Transfer     | ✔       | ✔     | ✔     |
| File Editor       | ✔       | ✔     | ✔     |
| Delete File       | ✔       | ✔     | ✔     |
| Code Highlighting | ✔       | ✔     | ✔     |
| Desktop Monitor   | ✔       | ✔     | ✔     |
| Screenshot        | ✔       | ✔     | ✔     |
| OS Info           | ✔       | ✔     | ✔     |
| Remote Terminal   | ✔       | ✔     | ✔     |
| Shellcode inject  | ✔       | ✔     | x     |
| Download and exec | ✔       | ✔     | x     |
| Keylogger         | ✔       | x     | x     |
| * Shutdown        | ✔       | ✔     | ✔     |
| * Reboot          | ✔       | ✔     | ✔     |


🚨 **Functions marked with * may require administrator/root privileges.**

---

## 🔧 Development

### Components
This project consists of three main components:
- **Client**
- **Server**
- **Front-end**

For OS support beyond Linux and Windows, additional C compilers may be required. For example, to support Android, install [Android NDK](https://developer.android.com/ndk/downloads).

### Build Guide

The frontend requires no npm or build step — all dependencies are vendored ESM bundles.

```bash
# Clone the repository
git clone https://github.com/ironargondev/FeArKit
cd ./FeArKit

# Build client and server (using Docker)
make client
make server

# Or build locally (requires Go 1.22+, statik)
go mod tidy
go mod download
./scripts/build.client.sh
./scripts/build.server.sh   # embeds web-vue3/ via statik, then builds the binary
```

### Frontend Development

The frontend can be developed without rebuilding the binary. Build the server once, then use the `-dev` flag:

```bash
# Run with live frontend from filesystem (changes visible on browser refresh)
./build/server/server_linux_amd64 -dev ./web-vue3 -salt yoursalt
```
---

## Dependencies

FeArKit contains many third-party open-source projects.

Back-end dependencies are listed in `go.mod`. Front-end dependencies are vendored under `web-vue3/vendor/`.

Some major dependencies are listed below.

### Back-end

* [Go](https://github.com/golang/go) ([License](https://github.com/golang/go/blob/master/LICENSE))

* [gin-gonic/gin](https://github.com/gin-gonic/gin) (MIT License)

* [imroc/req](https://github.com/imroc/req) (MIT License)

* [kbinani/screenshot](https://github.com/kbinani/screenshot) (MIT License)

* [shirou/gopsutil](https://github.com/shirou/gopsutil) ([License](https://github.com/shirou/gopsutil/blob/master/LICENSE))

* [gorilla/websocket](https://github.com/gorilla/websocket) (BSD-2-Clause License)

* [orcaman/concurrent-map](https://github.com/orcaman/concurrent-map) (MIT License)

### Front-end

* [Vue 3](https://github.com/vuejs/core) (MIT License)

* [Element Plus](https://github.com/element-plus/element-plus) (MIT License)

* [vue3-sfc-loader](https://github.com/FranckFreiburger/vue3-sfc-loader) (MIT License)

* [xterm.js](https://github.com/xtermjs/xterm.js) (MIT License)

* [CodeMirror 6](https://github.com/codemirror/dev) (MIT License)

* [zmodem.js](https://github.com/FranckFreiburger/zmodem.js) (MIT License)

### Acknowledgements

* [natpass](https://github.com/lwch/natpass) (MIT License)
* Image difference algorithm inspired by natpass.

---

## 📜 License

Distributed under the [BSD-2 License](./LICENSE).
