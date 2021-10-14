使用python脚本,包含加密和解密,包含V2和旧版。
```shell
python decSecucrt.py dec -v2 -p [pass] "xxxxxx"
```

使用decrypt.go解密文件，更新后会自动通过注册表找到默认配置文件路径。
```shell
go run decrypt.go [pass] [secucrt/xxx.ini]
```
