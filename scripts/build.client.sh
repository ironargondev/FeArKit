export GO111MODULE=auto
export COMMIT=`git rev-parse HEAD`

export GOOS=linux
export GOARCH=arm
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/linux_arm FeArKit/client
export GOARCH=386
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/linux_i386 FeArKit/client
export GOARCH=arm64
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/linux_arm64 FeArKit/client
export GOARCH=amd64
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/linux_amd64 FeArKit/client

export GOOS=windows
export GOARCH=386
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/windows_i386 FeArKit/client
export GOARCH=arm64
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/windows_arm64 FeArKit/client
export GOARCH=amd64
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/windows_amd64 FeArKit/client

export GOOS=freebsd
export GOARCH=386
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/freebsd_i386 FeArKit/client
export GOARCH=amd64
go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/freebsd_amd64 FeArKit/client

# export CGO_ENABLED=1
# export GOOS=android

# export GOARCH=arm
# export CC=armv7a-linux-androideabi21-clang
# export CXX=armv7a-linux-androideabi21-clang++
# go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/android_arm FeArKit/client

# export GOARCH=386
# export CC=i686-linux-android21-clang
# export CXX=i686-linux-android21-clang++
# go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/android_i386 FeArKit/client

# export GOARCH=arm64
# export CC=aarch64-linux-android21-clang
# export CXX=aarch64-linux-android21-clang++
# go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/android_arm64 FeArKit/client

# export GOARCH=amd64
# export CC=x86_64-linux-android21-clang
# export CXX=x86_64-linux-android21-clang++
# go build -ldflags "-s -w -X 'FeArKit/client/config.Commit=$COMMIT'" -o ./build/client/android_amd64 FeArKit/client
