Build static version

# Own system
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o hyprspace hyprspace.go

# Cross-compile to MIPS 74Kc V5.0
CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -a -ldflags '-extldflags "-static"' -o hyprspace.mips hyprspace.go