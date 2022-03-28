package main

import "C"
import "fmt"

func main() {
	fmt.Println(C.int(10))
}

// CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="zig cc -target x86_64-macos --sysroot=/Library/Developer/CommandLineTools/SDKs/MacOSX10.15.sdk/usr" go build --tags extended
// CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="zig cc -target x86_64-macos --sysroot=/Library/Developer/CommandLineTools/usr/lib/clang/12.0.0" go build --tags extended
// CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="zig cc -target x86_64-macos" CXX="zig c++ -target x86_64-macos" go build
///CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="zcc" CXX="zxx" go build --tags extended

// zig cc -target aarch64-macos --sysroot=/Library/Developer/CommandLineTools/usr/lib/clang/12.0.0 -o hello hello.c
// zig cc -target aarch64-macos --sysroot=/home/kubkon/macos-SDK -I/usr/include -L/usr/lib -F/System/Library/Frameworks -framework CoreFoundation -o hello hello.c
