Building gSuneido
================

The Windows version with DLL and COM support needs [cgo](https://golang.org/cmd/cgo/) so it requires a C/C++ compiler on the path. I am using [Mingw-w64](https://sourceforge.net/projects/mingw-w64/). You can either install Msys2 or just Mingw-w64. The specific version of GCC shouldn't be critical (but not too ancient). I would recommend using a normal install of Go, *not* installing it in Msys2 with pacman. Go just needs gcc on the path.

From the terminal, running:

    g++ --version
    
Should give something like:

    g++ (x86_64-posix-seh-rev0, Built by MinGW-W64 project) 8.1.0

I've been mostly working on the Windows version so building on Linux or Mac may occasionally be broken due to build tags etc. Cgo is not required for these versions, they are plain Go. But the only UI is the command line REPL.

I use the included makefile (requires make). If you prefer not to use make you can just look at the makefile to see what the build commands are. There is a make included with mingw_w64 but it's called `mingw32_make.exe` You can make a copy and rename it to `make.exe`

You should be able to run the tests with the usual: `go test ./..`

To run the [portable tests](https://github.com/apmckinlay/suneido_tests) (shared with cSuneido and jSuneido) you need to clone or download the separate Github repository. It needs to be a peer of the gsuneido directory, and must be called suneido_tests. e.g. `c:/stuff/mygsuneido` and `c:/stuff/suneido_tests`

All my development and testing has been on 64 bit. In theory you could build a 32 bit version of gSuneido but I suspect it would require some changes. It might be useful for client mode since it would use less memory because pointers and int's would be smaller (32 bits instead of 64) The client is unlikely to need access to more than 2gb of memory.

There are some warnings for the code - `composite literal uses unkeyed fields` and `possible misuse of unsafe.Pointer`. I've been trying to fix other warnings.

My normal environment is [Visual Studio Code](https://code.visualstudio.com/) with the [Go extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode.Go) using gopls. Debugging (with the included Delve) can be problematic, sometimes it works.
