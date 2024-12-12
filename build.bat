@go install github.com/tc-hib/go-winres@latest
@go-winres simply --icon eqssr.png
@go build -trimpath -buildmode=pie -ldflags="-s -w" -o eqssr.exe main.go

go install github.com/akavel/rsrc@latest
rsrc -ico eqssr.ico
go build -trimpath -buildmode=pie -ldflags="-s -w"
move eqssr.exe bin\eqssr.exe