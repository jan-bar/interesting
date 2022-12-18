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

func List(s string, pass ...string) ([]FileList, error) {
	cmd := &exec.Cmd{
		Path: z7path, // 已经通过init成功,这里直接返回对象
		Args: append([]string{z7path}, "l", "-slt", s),
	}
	if len(pass) > 0 {
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
		_ = cmd.Wait() // 保证子进程退出
	}()

	var (
		tmp FileList
		ret []FileList
	)

	br := bufio.NewScanner(cr)
	for br.Scan() {
		line := strings.Split(br.Text(), " = ")
		if len(line) != 2 {
			continue
		}

		switch line[0] {
		case "Path":
			if tmp.Path != "" {
				ret = append(ret, tmp)
			}
			tmp.Path = line[1]
		case "Size":
			if line[1] == "" {
				tmp.Size = 0
			} else {
				tmp.Size, err = strconv.ParseInt(line[1], 10, 64)
				if err != nil {
					return nil, err
				}
			}
		case "Packed Size":
			if line[1] == "" {
				tmp.PackedSize = 0
			} else {
				tmp.PackedSize, err = strconv.ParseInt(line[1], 10, 64)
				if err != nil {
					return nil, err
				}
			}
		case "Modified":
			tmp.Modified = line[1]
		case "Attributes":
			tmp.Attributes = line[1]
		case "CRC":
			tmp.CRC = line[1]
		}
	}
	return ret, br.Err()
}
