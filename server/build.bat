SETLOCAL

set CGO_ENABLED=1
set PATH=%PATH%;C:\msys64\mingw64\bin

go build -tags=jsoniter .
my-desktop-server
