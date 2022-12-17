//go:build windows

package exec

import (
	_ "embed"
)

//go:embed 7za.exe
var z7Data []byte

// windows 可能的7z命令
var z7Name = []string{"7z.exe", "7za.exe"}

/*
https://7-zip.org/download.html

下载window控制台程序,x64\7za.exe
https://7-zip.org/a/7z2201-extra.7z
*/
