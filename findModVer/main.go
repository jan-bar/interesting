package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func main() {
	if len(os.Args) > 1 {
		err := test(os.Args[1])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func test(path string) error {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = path // 分析该目录的模块

	cr, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	br := bufio.NewScanner(cr)
	br.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '}'); i >= 0 {
			return i + 1, data[:i+1], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	var (
		md struct {
			GoMod string
		}
		list []fileVer
	)
	for br.Scan() {
		err = json.Unmarshal(br.Bytes(), &md)
		if err != nil {
			continue
		}

		err = parseModFile(md.GoMod, &list)
		if err != nil {
			return err
		}
	}
	err = br.Err()
	if err != nil {
		return err
	}

	// copy module.Sort
	sort.Slice(list, func(i, j int) bool {
		mi := list[i]
		mj := list[j]
		if mi.ver.Path != mj.ver.Path {
			return mi.ver.Path < mj.ver.Path
		}
		vi := mi.ver.Version
		vj := mj.ver.Version
		var fi, fj string
		if k := strings.Index(vi, "/"); k >= 0 {
			vi, fi = vi[:k], vi[k:]
		}
		if k := strings.Index(vj, "/"); k >= 0 {
			vj, fj = vj[:k], vj[k:]
		}
		if vi != vj {
			return semver.Compare(vi, vj) < 0
		}
		return fi < fj
	})

	j := 0
	tmp := list[j].ver.Path
	for i := 1; i < len(list); i++ {
		if tmp == list[i].ver.Path {
			continue
		}

		var show []string // 去重
		dup := make(map[string]struct{}, i-j)
		for k := j; k < i; k++ {
			s := fmt.Sprintf("    %s -> %s -> %s",
				list[k].ver.Version, list[k].tp, list[k].file)

			if _, ok := dup[s]; !ok {
				show = append(show, s)
				dup[s] = struct{}{}
			}
		}

		fmt.Printf("\n\npath: %s\n", tmp)
		for _, v := range show {
			fmt.Println(v)
		}

		j = i
		tmp = list[j].ver.Path
	}

	return nil
}

type fileVer struct {
	ver  module.Version
	file string
	tp   string
}

func parseModFile(s string, ver *[]fileVer) error {
	data, err := os.ReadFile(s)
	if err != nil {
		return err
	}

	mf, err := modfile.Parse(s, data, nil)
	if err != nil {
		return err
	}

	for _, require := range mf.Require {
		*ver = append(*ver, fileVer{
			ver:  require.Mod,
			file: s,
			tp:   "require indirect=" + strconv.FormatBool(require.Indirect),
		})
	}

	for _, exclude := range mf.Exclude {
		*ver = append(*ver, fileVer{
			ver:  exclude.Mod,
			file: s,
			tp:   "exclude",
		})
	}

	for _, replace := range mf.Replace {
		*ver = append(*ver, fileVer{
			ver:  replace.Old,
			file: s,
			tp:   "replace old",
		})
		*ver = append(*ver, fileVer{
			ver:  replace.New,
			file: s,
			tp:   "replace new",
		})
	}

	return nil
}
