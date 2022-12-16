package z7

/* 查询的各种资料
https://git.zx2c4.com/wireguard-windows/tree/elevate/shellexecute.go#n128
https://github.com/harvimt/pylib7zip
https://gist.github.com/harvimt/9461046
https://github.com/adoconnection/SevenZipExtractor
https://www.anquanke.com/post/id/201221
https://stackoverflow.com/questions/37781676/how-to-use-com-component-object-model-in-golang
https://github.com/go-ole/go-ole

主要是C++导出的dll和C导出的dll是有差异的
该项目展示了调用7z的C版本生成的dll用法: https://github.com/itchio/sevenzip-go

至于C++的dll,多了一层*Vtable指针,并且每个方法调用时第一个参数必须是自身
就像对象调用方法时,该方法里面可以直接访问对象内容,并且 IUnknown 对象也在C++继承时提现

本项目支持:
  windows && !cgo, 直接通过syscall系统调用dll
  (windows || linux) && cgo, 通过C方式调用dll
*/
