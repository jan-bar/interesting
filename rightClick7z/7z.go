//go:build windows
// +build windows

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var isDepth bool

func main() {
	if len(os.Args) != 4 {
		msgBox(fmt.Sprintf(`Usage:%s  7zipPath  srcFile  dstDir`, os.Args[0]))
		return
	}

	err := exec7z(os.Args[1], os.Args[2], os.Args[3], bufio.NewReader(nil))
	if err != nil {
		msgBox(err.Error())
	}
}

func msgBox(s string) {
	msg, _ := syscall.UTF16PtrFromString(s)
	info, _ := syscall.UTF16PtrFromString("Error")
	_, _, _ = syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").
		Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(info)), 16)
}

func inputBox(title, message string) (string, error) {
	// https://github.com/martinlindhe/inputbox 参考该项目
	cmd := exec.Command("powershell.exe", "-NoExit", "-Command", "-")
	cmd.Stdin = strings.NewReader(`
[Console]::OutputEncoding=[System.Text.Encoding]::UTF8
[void][Reflection.Assembly]::LoadWithPartialName('Microsoft.VisualBasic')
$title='` + title + `'
$msg='` + message + `'
$answer=[Microsoft.VisualBasic.Interaction]::InputBox($msg, $title)
Write-Output $answer;`)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run() // 点击取消时结果为空,不输入时结果也为空
	return strings.TrimSpace(out.String()), err
}

const (
	spCharLen = 72 // 压缩包内容被夹在这两个'-'中间
	pathStart = 53 // 文件名开始位置
)

// http://www.bandisoft.com/bandizip/help/auto_dest/
// 自动解压原理,只有一个文件则解压到当前文件,都在一个目录中则解压到这个目录
// 否则创建一个文件名的目录解压到里面去
func exec7z(p7z, srcFile, dstDir string, br *bufio.Reader) error {
	fi, err := os.Stat(srcFile)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		if isDepth {
			return nil // 递归解压遇上目录返回正常
		}
		return fmt.Errorf("不能解压目录: %s", srcFile)
	}

	pass, cmd, err := startCmd(filepath.Join(p7z, "7z.exe"), srcFile, br)
	if err != nil {
		if isDepth {
			return nil // 递归解压遇上遍历文件失败
		}
		return err
	}

	var (
		root     string
		multiple bool
	)
	for {
		line, err := readLineSlice(br)
		if err != nil {
			return err
		}
		if len(line) < pathStart {
			continue // 跳过不合法的行
		}

		if strings.Count(line, "-") == spCharLen {
			break // 再次遇到该标志结束
		}

		line = line[pathStart:] // 去掉前面固定长度字符
		if i := strings.IndexByte(line, os.PathSeparator); i > 0 {
			line = line[:i] // 得到一级目录或文件名
		}

		if root != line {
			if root != "" {
				multiple = true // 说明存在不同顶级目录
				break
			}

			r := make([]byte, len(line))
			copy(r, line) // 字符串的深复制,因为line复用[]byte缓存
			root = *(*string)(unsafe.Pointer(&r))
		}
	}
	//goland:noinspection GoUnhandledErrorResult
	go cmd.Process.Kill() // 到这里可以杀死子进程

	if multiple {
		dstDir = filepath.Join(dstDir, filepath.Base(srcFile))
	}

	var arg []string
	if pass == "" {
		arg = []string{"x", "-o" + dstDir, srcFile}
	} else {
		arg = []string{"x", "-p" + pass, "-o" + dstDir, srcFile}
	}
	// 调用7zG.exe执行解压,目标文件不存在会自动创建
	err = exec.Command(filepath.Join(p7z, "7zG.exe"), arg...).Run()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 255 {
			return nil // 用户取消操作不弹窗提示,其他错误由7zG.exe弹窗
		}
		return err
	}

	if multiple {
		return nil
	}
	isDepth = true // 单个文件递归解压,直到内层无法解压,内层解压有些报错可忽略
	return exec7z(p7z, filepath.Join(dstDir, root), dstDir, br)
}

func startCmd(p7z, srcFile string, br *bufio.Reader) (string, *exec.Cmd, error) {
	var pass string
	cmd := exec.Command(p7z, "l", srcFile)
usePass:
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cr, err := cmd.StdoutPipe()
	if err != nil {
		return "", nil, err
	}
	if err = cmd.Start(); err != nil {
		return "", nil, err
	}

	br.Reset(cr)
	for {
		line, err := readLineSlice(br)
		if err != nil {
			return "", nil, fmt.Errorf("%s list失败！[%w]", srcFile, err)
		}
		if len(line) <= 0 {
			continue
		}

		if strings.Count(line, "-") == spCharLen {
			return pass, cmd, nil // 遇到该标志才开始遍历
		}

		// 首次提示输入密码 || 密码错误需要重试
		if strings.HasPrefix(line, "Enter password") ||
			(pass != "" && strings.HasPrefix(line, "Errors: 1")) {
			for {
				pass, err = inputBox("Enter password", srcFile)
				if err != nil {
					return "", nil, err
				}
				if pass != "" {
					break // 输入为空或点击取消,重试直到获取到输入
				}
			}

			//goland:noinspection GoUnhandledErrorResult
			go cmd.Process.Kill()
			// 携带密码再试一次
			cmd = exec.Command(p7z, "l", "-p"+pass, srcFile)
			goto usePass
		}
	}
}

func readLineSlice(br *bufio.Reader) (string, error) {
	var line []byte
	for {
		l, more, err := br.ReadLine()
		if err != nil {
			return "", err
		}
		if line == nil && !more {
			// 注意这里返回的string会因为byte内容改变而改变
			return *(*string)(unsafe.Pointer(&l)), nil
		}
		line = append(line, l...)
		if !more {
			return *(*string)(unsafe.Pointer(&line)), nil
		}
	}
}
