export GO111MODULE=auto
export COMMIT=`git rev-parse HEAD`

# patch the web
cd ./web && npm install && npm run build-prod

cd .. && sudo GOBIN=/usr/local/bin/ go install github.com/rakyll/statik

/usr/local/bin/statik -m -src="./web/dist" -f -dest="./server/embed" -p web -ns web

go mod tidy

#export GOOS=darwin
#export GOARCH=arm64
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_darwin_arm64 FeArKit/server
#export GOARCH=amd64
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_darwin_amd64 FeArKit/server

export GOOS=linux
#export GOARCH=arm
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_linux_arm FeArKit/server
#export GOARCH=386
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_linux_i386 FeArKit/server
#export GOARCH=arm64
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_linux_arm64 FeArKit/server
export GOARCH=amd64
go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_linux_amd64 FeArKit/server

#export GOOS=windows
#export GOARCH=386
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_windows_i386.exe FeArKit/server
#export GOARCH=arm64
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_windows_arm64.exe FeArKit/server
#export GOARCH=amd64
#go build -ldflags "-s -w -X 'FeArKit/server/config.Commit=$COMMIT'" -tags=jsoniter -o ./build/server/server_windows_amd64.exe FeArKit/server
