package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
			break // 从下面开始遍历
		}
	}

	var tmp string
	for {
		line, err := readLineSlice(br)
		if err != nil {
			return err
		}
		if strings.HasPrefix(line, "Path = ") {
			if i := strings.IndexByte(line, os.PathSeparator); i > 0 {
				line = line[:i]
			}
			fmt.Println(line)
			if tmp == "" {
				tmp = line
			} else if tmp != line {
				fmt.Println(tmp, line, "have")
				break
			}
		}
	}

	//return exec7zG(p7z, srcFile, dstDir)
	return nil
}

func exec7zG(p7z, srcFile, dstDir string) error {
	exe7zG := filepath.Join(p7z, "7zG.exe")
	return exec.Command(exe7zG, "x", "-o"+dstDir, srcFile).Run()
}

func readLineSlice(br *bufio.Reader) (string, error) {
	var line []byte
	for {
		l, more, err := br.ReadLine()
		if err != nil {
			return "", err
		}
		if line == nil && !more {
			return *(*string)(unsafe.Pointer(&l)), nil
		}
		line = append(line, l...)
		if !more {
			return *(*string)(unsafe.Pointer(&line)), nil
		}
	}
}
