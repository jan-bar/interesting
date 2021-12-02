package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {
	err := set7zIco()
	if err != nil {
		fmt.Println(err)
	}
}

func set7zIco() error {
	haveIco, err := createIco()
	if err != nil {
		return err
	}
	regKey, err := get7zKeys()
	if err != nil {
		return err
	}

	for k, v := range regKey {
		path, ok := haveIco[k]
		if !ok {
			path = haveIco["default"]
		}
		err = set7zDefaultIcon(v, path)
		if err != nil {
			return err
		}
	}
	fmt.Printf("reg len=%d\n", len(regKey))

	// 关闭资源管理器
	err = exec.Command("taskkill", "/f", "/im", "explorer.exe").Run()
	if err != nil {
		return err
	}
	// 运行资源管理器
	err = exec.Command("cmd", "/c", "start", "explorer.exe").Run()
	if err != nil {
		return err
	}
	return nil
}

func createIco() (map[string]string, error) {
	fs, err := os.ReadDir("ico")
	if err != nil {
		return nil, err
	}
	ico := make(map[string]string)
	for _, v := range fs {

	}
	return nil, nil
}

func set7zDefaultIcon(key, val string) error {
	key = key + "\\DefaultIcon"
	k, err := registry.OpenKey(registry.CLASSES_ROOT, key, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer k.Close()
	fmt.Printf("%s\t%s\n", key, val)
	return k.SetStringValue("", val)
}

func get7zKeys() (map[string]string, error) {
	k, err := registry.OpenKey(registry.CLASSES_ROOT, "", registry.ALL_ACCESS)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer k.Close()

	sub, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for _, v := range sub {
		if strings.HasPrefix(v, "7-Zip.") {
			ret[v[6:]] = v
		}
	}
	return ret, nil
}

func fileNotExists(f string) bool {
	_, err := os.Stat(f)
	return os.IsNotExist(err)
}
