//go:build linux

package exec

import (
	_ "embed"
)

//go:embed 7zzs
var z7Data []byte

// Linux 可能的7z命令
var z7Name = []string{"7z", "7za", "7zz", "7zzs"}

/*
https://7-zip.org/download.html

下载Linux可执行程序,7zzs是没有依赖的可执行程序
https://7-zip.org/a/7z2201-linux-x64.tar.xz
*/
