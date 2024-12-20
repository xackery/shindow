:: go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
goversioninfo -icon=ico/icon.ico -manifest=shindow.exe.manifest -o=rsrc.syso versioninfo.json
go build -ldflags "-H windowsgui -X main.Version=1.2.1.1" || exit /b
move shindow.exe bin/shindow.exe || exit /b
cd bin || exit /b
shindow.exe c:\games\eq\thj\Logs\eqlog_Shin_thj.txt || exit /b