SETLOCAL

set CGO_ENABLED=1
set PATH=%PATH%;C:\msys64\mingw64\bin

go build -tags=jsoniter .
if %errorlevel% neq 0 exit /b %errorlevel%
go build ./services/email-injestor
if %errorlevel% neq 0 exit /b %errorlevel%
nodemon
my-desktop-server
