export GO111MODULE=auto
export COMMIT=`git rev-parse HEAD`


export GOOS=linux
export GOARCH=amd64
go build -ldflags "-s -w -X 'FeArKit/tools/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/tools/tools_linux_amd64 FeArKit/tools

export GOOS=windows
export GOARCH=amd64
go build -ldflags "-s -w -X 'FeArKit/tools/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/tools/tools_windows_amd64.exe FeArKit/tools
