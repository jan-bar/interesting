package main

import (
	"fmt"

	"interesting/zip7/exec"
)

func main() {
	f, err := exec.List(`E:\360xiazai\myGit.7z`)
	if err != nil {
		panic(err)
	}
	for _, v := range f {
		fmt.Printf("%+v\n", v)
	}
}
