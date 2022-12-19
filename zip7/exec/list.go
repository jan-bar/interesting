package exec

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
)

type FileList struct {
	Path       string
	Type       string
	Size       int64
	PackedSize int64
	Modified   string
	Attributes string
	CRC        string
}

func List(s string, pass ...string) (fs []FileList, err error) {
	cmd := &exec.Cmd{
		Path: z7path, // 已经通过init成功,这里直接返回对象
		Args: append([]string{z7path}, "l", "-slt", s),
	}
	if len(pass) > 0 && pass[0] != "" {
		cmd.Args = append(cmd.Args, "-p"+pass[0])
	}

	cr, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	defer func() {
		err1 := cmd.Wait() // 保证子进程退出
		if err == nil && err1 != nil {
			if e, ok := err1.(*exec.ExitError); ok {
				err = Z7Err(e.ExitCode())
			}
		}
	}()

	var (
		tmp FileList
		ret []FileList
	)

	br := bufio.NewScanner(cr)
	for br.Scan() {
		key, val, ok := strings.Cut(br.Text(), " = ")
		if !ok {
			continue
		}

		switch key {
		case "Path":
			if tmp.Path != "" {
				ret = append(ret, tmp)
			}
			tmp.Path = val
		case "Size":
			if val == "" {
				tmp.Size = 0
			} else {
				tmp.Size, err = strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, err
				}
			}
		case "Packed Size":
			if val == "" {
				tmp.PackedSize = 0
			} else {
				tmp.PackedSize, err = strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, err
				}
			}
		case "Modified":
			tmp.Modified = val
		case "Attributes":
			tmp.Attributes = val
		case "CRC":
			tmp.CRC = val
		}
	}
	return ret, br.Err()
}
