//go:build windows
// +build windows

package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func main() {
	if len(os.Args) != 4 {
		return
	}
	err := exec7z(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		println(err.Error())
	}
}

func exec7z(p7z, srcFile, dstDir string) error {
	exe7z := filepath.Join(p7z, "7z.exe")
	cmd := exec.Command(exe7z, "l", "-slt", srcFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cr, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}

	br := bufio.NewReader(cr)
	for {
		line, err := readLineSlice(br)
		if err != nil {
			return err
		}
		if line == "----------" {
			break // 遇到该标志才开始遍历
		}
	}

	var (
		root     string
		multiple bool
	)
	for {
		line, err := readLineSlice(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if len(line) > 7 && strings.EqualFold(line[:7], "Path = ") {
			if i := strings.IndexByte(line, os.PathSeparator); i > 0 {
				line = line[7:i]
			} else {
				line = line[7:]
			}

			if root != line {
				if root != "" {
					//goland:noinspection GoUnhandledErrorResult
					go cmd.Process.Kill()
					multiple = true // 说明存在不同顶级目录
					break
				}

				r := make([]byte, len(line))
				copy(r, line) // 字符串的深复制
				root = *(*string)(unsafe.Pointer(&r))
			}
		}
	}

	if multiple {
		dstDir = filepath.Join(dstDir, filepath.Base(srcFile))
	}
	return exec.Command(filepath.Join(p7z, "7zG.exe"),
		"x", "-o"+dstDir, srcFile).Run()
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
