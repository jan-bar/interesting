package exec

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

var z7path string

func init() {
	var (
		buf    bytes.Buffer
		temp7z = filepath.Join(os.TempDir(), z7Name[0])
		head7z = []byte("7-Zip")
	)
	for _, v := range append(z7Name, temp7z) {
		cmd := exec.Command(v)
		if cmd.Err != nil {
			continue
		}

		buf.Reset()
		cmd.Stdout = &buf
		err := cmd.Run()
		if err == nil && bytes.Contains(buf.Bytes(), head7z) {
			z7path = cmd.Path
			break // 找到一个可用的7z可执行程序
		}
	}

	if z7path == "" {
		// 每次都尝试系统中找到7z可执行程序,找不到才将嵌入数据写入临时文件
		// 如果系统一直没有安装7z,则后面每次都使用临时7z文件
		// 如果后续系统安装了7z,那么会继续使用系统中的7z可执行程序
		// 确保目标window和Linux系统,一定可以执行成功7z命令
		fw, err := os.OpenFile(temp7z,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			panic(err)
		}
		_, err = fw.Write(z7Data)
		_ = fw.Close()
		if err != nil {
			panic(err)
		}
	}
}

// Path7z 确保外部不可更改7z可执行程序路径
func Path7z() string { return z7path }

// Command 直接返回 *exec.Cmd 对象
func Command(arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: z7path, // 已经通过init成功,这里直接返回对象
		Args: append([]string{z7path}, arg...),
	}
}
