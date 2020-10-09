#!/bin/sh
go build -o build/goraft-cli/goraft-cli cmd/goraft-cli/main.go
go build -o build/goraft-daemon/goraft-daemon cmd/goraft-daemon/main.go