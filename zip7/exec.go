package main

import (
	"flag"
	"fmt"

	"interesting/zip7/exec"
)

func main() {
	fs := flag.String("f", "", "archive file")
	pass := flag.String("p", "", "password")
	mode := flag.String("m", "", "l list,a add,e extract")
	flag.Parse()

	// 通过调用7z命令行工具,完成列表、加密、解密相关逻辑
	switch *mode {
	case "l":
		f, err := exec.List(*fs, *pass)
		if err != nil {
			panic(err)
		}
		for _, v := range f {
			fmt.Printf("%+v\n", v)
		}
	case "a":
		arg := make([]string, 2, 4+flag.NArg())
		arg[0] = "a"
		arg[1] = *fs
		if *pass != "" {
			// 加密并且加密文件头部
			arg = append(arg, "-p"+*pass, "-mhe")
		}
		arg = append(arg, flag.Args()...)
		cmd := exec.Command(arg...)
		err := cmd.Run()
		fmt.Println(cmd.Args) // 打印完整命令
		if err != nil {
			panic(err)
		}
	case "e":
		arg := make([]string, 2, 5)
		arg[0] = "e"
		arg[1] = *fs
		if *pass != "" {
			// 加密并且加密文件头部
			arg = append(arg, "-p"+*pass)
		}
		// 强制覆盖不提示
		arg = append(arg, "-aoa", "-o"+*fs+"_dec")
		cmd := exec.Command(arg...)
		err := cmd.Run()
		fmt.Println(cmd.Args) // 打印完整命令
		if err != nil {
			panic(err)
		}
	}
}
