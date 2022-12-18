package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

/*
https://7-zip.org/download.html

下载window控制台程序,x64\7za.exe
https://7-zip.org/a/7z2201-extra.7z

中文帮助文档: https://github.com/sparanoid/7z/blob/master/zh-cn/7-zip.chm

7-Zip (a) 22.01 (x64) : Copyright (c) 1999-2022 Igor Pavlov : 2022-07-15

Usage: 7za <command> [<switches>...] <archive_name> [<file_names>...] [@listfile]

<Commands>
  a : 添加文件到压缩档案
  b : 测试 CPU 运行速度及检查内存错误
  d : 从压缩档案删除文件
  e : 从存档中提取文件（不使用目录名）
  h : 计算文件的哈希值
  i : 显示有关支持格式的信息
  l : 列出存档内容
 rn : 重命名存档中的文件
  t : 测试存档的完整性
  u : 更新文件以存档
  x : 使用完整路径提取文件

<Switches>
  -- : 在命令行中使"--"后的选项开关"-"都失效。这样就允许在命令行中使用文件名以"-"开头的文件
  -ai[r[-|0]]{@listfile|!wildcard} : 指定附加文件，包括压缩包文件名及通配符
  -ax[r[-|0]]{@listfile|!wildcard} : 指定必须从操作中排除的压缩包
  -ao{a|s|t|u} : 指定在释放期间如何覆盖硬盘上现有的同名文件
      a: 直接覆盖现有文件，而没有任何提示
      s: 跳过现有文件，其不会被覆盖
      t: 如果相同文件名的文件以存在，将自动重命名现有的文件。举个例子，文件 file.txt 将被自动重命名为 file_1.txt
      u: 如果相同文件名的文件以存在，将自动重命名被释放的文件。举个例子，文件 file.txt 将被自动重命名为 file_1.txt
  -an : 不解析命令行中的 archive_name 区域, 此选项必须和 -i (附加文件) 开关一起使用
  -bb[0-3] : 设置输出日志级别
      0: 禁用(默认)
      1: 可以直接 -bb ,在日志中显示已处理文件的名称
      2: 显示在固体存档中内部处理的其他文件的名称：“提取”操作跳过的文件，“添加”/“更新”操作重新打包的文件。
      3: 显示有关“添加”/“更新”操作的附加操作（分析、复制）的信息。
  -bd : 禁用进度指示器
  -bs{o|e|p}{0|1|2} : 为 output/error/progress 行设置输出流
      o: 标准输出信息
      e: 错误信息
      p: 进度信息
      0: 禁用流
      1: 重定向到标准输出
      2: 重定向到标准错误
  -bt : 显示执行时间统计
  -i[r[-|0]]{@listfile|!wildcard} : 包括文件名
     r[-|0]: 和后面递归-r相同
     @listfile: 列表文件
     !wildcard: 通配符
  -m{Parameters} : 设置压缩方式
    -mmt[N] : 设置 CPU 线程数：[off | on | {N}]，默认on，可指定线程数
    -mx[N] : 设置压缩级别：0仅存储,1急速压缩,3快速压缩,5标准压缩 (默认),7最大压缩,9极限压缩
    zip:
        -mm{MethodID}: 设置方法：Copy、Deflate (默认)、Deflate64、BZip2、LZMA、PPMd。
        -mfb{NumFastBytes}: 设置 Deflate 编码器的快速字节数。(3,258) 默认32
        -mpass{NumPasses}: 设置 Deflate 编码器的遍数。默认1
        -md{Size}[b|k|m]: 设置bzip2字典大小, 默认900000
        -mmem{Size}[b|k|m]: 设置用于 PPMd 的内存大小，默认24
        -mo{Size}: 设置 PPMd 的模型顺序。大小必须在 [2,16] 范围内。默认值为 8
        -mem{EncryptionMethodID}: 设置加密方法：ZipCrypto (默认)、AES128、AES192、AES256
        -mcl[off | on]: 7-Zip 始终使用本地代码页作为文件名。默认off
        -mcu[off | on]: 7-Zip 对包含非 ASCII 符号的文件名使用 UTF-8。默认off
        -mcp{CodePage}: 设置代码页。默认off
        -mtm[off | on]: 存储文件的最后修改时间戳。默认on
        -mtc[off | on]: 存储文件的创建时间戳。默认off
        -mta[off | on]: 存储文件的最后访问时间戳。默认off
        -mtp{N}: 设置时间戳精度：0 - Windows（100 ns），1 - Unix（1 秒），2 - DOS（2 秒）。 3 - Windows（100 纳秒）。默认0
    gzip:
        GZip 使用与 Zip 相同的参数，但 GZip 仅使用 Deflate 方法进行压缩。所以 GZip 只支持以下参数：x、fb、pass。
    bzip2:
        -mpass{NumPasses}: 设置 Bzip2 编码器的遍数。默认1
        -md{Size}[b|k|m]: 设置bzip2字典大小, 默认900000
    7z:
        -mx[N]
            压缩等级 压缩算法 字典大小 快速字节 匹配器 过滤器 描述
            0        Copy                                 无压缩
            1        LZMA      64 KB   32    HC4    BCJ   最快压缩
            3        LZMA      1 MB    32    HC4    BCJ   快速压缩
            5        LZMA      16 MB   32    BT4    BCJ   正常压缩
            7        LZMA      32 MB   64    BT4    BCJ   最大压缩
            9        LZMA      64 MB   64    BT4    BCJ2  极限压缩
        -myx[N]:
            Level Description
            0 No analysis.
            1 or more WAV file analysis (for Delta filter).
            7 or more EXE file analysis (for Executable filters).
            9 or more analysis of all files (Delta and executable filters).
  -o{Directory} : 设置输出目录,不存在会自动创建
  -p{Password} : 设置密码
  -r[-|0] : 用于名称搜索的递归子目录
     -r:  开启递归子目录。对于e(释放)、l(列表)、t(测试)、x(完整路径释放) 这些在压缩包中操作的命令，会默认使用此选项。
     -r-: 关闭递归子目录。对于a(添加)、d(删除)、u(更新)等所有需扫描磁盘文件的命令，会默认使用此选项。
     -r0: 开启递归子目录。但只应用于通配符。
  -sa{a|e|s} : 设置存档名称模式
      -saa: 始终添加存档类型扩展名。
      -sae: 使用命令中指定的确切名称。
      -sas: 仅当指定名称没有扩展名时才添加扩展名。这是默认选项。
  -scc{UTF-8|WIN|DOS} : 为控制台输入输出设置字符集
       UTF-8: Unicode UTF-8 编码集
       WIN:   Windows 默认编码集
       DOS:   Windows DOS (OEM) 编码集
  -scs{UTF-8|UTF-16LE|UTF-16BE|WIN|DOS|{id}} : 为列表文件设置字符集
       UTF-8:    Unicode UTF-8 编码集
       UTF-16LE: Unicode UTF-16 小端字符集。
       UTF-16BE: Unicode UTF-16 大端字符集。
       WIN:      Windows 默认编码集
       DOS:      Windows DOS (OEM) 编码集
       {id}:     代码页编号（在 Microsoft Windows 中指定）。
  -scrc[CRC32|CRC64|SHA1|SHA256|*] : 为 x、e、h 命令设置哈希函数
        *: 全都要
  -sdel : 压缩后删除文件(只有成功才删除)
  -seml[.] : 通过电子邮件发送存档
        .: 使用点会删除文件
  -sfx[{name}] : 创建 SFX 存档,以下模式文件必须和7z可执行程序同目录
       7z.sfx: 创建GUI方式的自解压
       7zCon.sfx: 创建控制台方式的自解压
     上述2个文件可以在7z安装包中找到，还可以用copy命令创建自解压安装文件
  -si[{name}] : 从标准输入读取数据
      name: 为标准输入内容命名,没有则文件没名字
  -slp : 设置大页面模式
     -slp:  开启大页面
     -slp-: 关闭大内存模式。此选项为默认值。
  -slt : 显示 l (List) 命令的技术信息
     会按照每个文件显示详细信息,而不是每行显示那种方式
  -snh : 将硬链接存储为链接
  -snl : 将符号链接存储为链接
  -sni : 存储 NT 安全信息，存储NTFS系统信息
  -sns[-] : 存储 NTFS 备用流
    -sns: 启用“存储 NTFS 备用流”模式。如果您提取存档，这是默认选项。
    -sns-: 禁用“存储 NTFS 备用流”模式。这是默认选项，如果您创建存档或调用“列表”命令。
  -so : 将数据写入标准输出
  -spd : 禁用文件名的通配符匹配
  -spe : 消除提取命令的根文件夹重复
  -spf : 使用完全限定的文件路径
    -spf:  使用绝对路径，包括驱动器号。
    -spf2: 使用没有驱动器号的完整路径。
  -ssc[-] : 设置区分大小写模式
    -ssc:  设置为区分大小写模式。Posix 及 Linux 系统默认使用此选项。
    -ssc-: 设置为不区分大小写模式。Windows 系统默认使用此选项。
  -sse : 如果无法打开某些输入文件，则停止创建存档
  -ssp : 归档时不要更改源文件的上次访问时间
  -ssw : 压缩共享文件，压缩正在被其他应用程序使用的文件
  -stl : 从最近修改的文件设置存档时间戳
  -stm{HexMask} : 设置 CPU 线程关联掩码（十六进制数）
  -stx{Type} : 排除存档类型,禁用指定格式打开文件
  -t{Type} : 设置存档类型，zip、7z、rar、cab、gzip、bzip2、tar，7z为默认值
  -u[-][p#][q#][r#][x#][y#][z#][!newArchiveName] : 更新选项
    -u-: 不进行任何更新
    -u!newArchiveName: 指定新压缩包的路径。
    -u<state><action>: 每个文件名都会赋予下列六个变量<state>
       <state>            状态说明                      磁盘上的文件  压缩包中的文件
          p    文件在压缩包中，但并不和磁盘上的文件相匹配。               存在，但并不匹配
          q    文件在压缩包中，但磁盘上并不存在。            不存在       存在
          r    文件不在压缩包中，但磁盘上存在。              存在         不存在
          x    压缩包中的文件比磁盘上的文件新。              较旧         较新
          y    压缩包中的文件比磁盘上的文件旧。              较新         较旧
          z    压缩包中的文件和磁盘上的文件相同。            相同         相同
          w    不能检测文件是否较新(时间相同但大小不同)        ?           ?
       <action> 为适当的 <state> 指定动作。
           <action> ::= 0 | 1 | 2 | 3
           0 忽略文件(在压缩包中不为此文件创建项目)
           1 复制文件(用压缩包中的新文件覆盖旧文件)
           2 压缩文件(将磁盘上的新文件压缩到档案中)
           3 创建剔除项(释放过程中将删除文件或目录项)。此功能只支持 7z 格式。
      下列表格中显示的是更新命令的动作设置
         命令<state> p q r x y z w
         d (删除)    1 0 0 0 0 0 0
         a (添加)    1 1 2 2 2 2 2
         u (更新)    1 1 2 1 2 1 2
  -v{Size}[b|k|m|g] : 创建卷,指定分卷大小,不指定单位按字处理
  -w[{path}] : 分配工作目录。空路径表示临时目录
  -x[r[-|0]]{@listfile|!wildcard} : 排除文件名
     使用指定递归方式,后面匹配列表文件和通配符
  -y : 对所有查询都假设是
*/

var z7path string

func init() {
	var (
		buf    bytes.Buffer
		temp7z = filepath.Join(os.TempDir(), z7Name[0])
		head7z = []byte("7-Zip")
	)
	for _, v := range append(z7Name, temp7z) {
		cmd := exec.Command(v)
		if cmd.Err != nil {
			continue
		}

		buf.Reset()
		cmd.Stdout = &buf
		err := cmd.Run()
		if err == nil && bytes.Contains(buf.Bytes(), head7z) {
			z7path = cmd.Path
			break // 找到一个可用的7z可执行程序
		}
	}

	if z7path == "" {
		// 每次都尝试系统中找到7z可执行程序,找不到才将嵌入数据写入临时文件
		// 如果系统一直没有安装7z,则后面每次都使用临时7z文件
		// 如果后续系统安装了7z,那么会继续使用系统中的7z可执行程序
		// 确保目标window和Linux系统,一定可以执行成功7z命令
		fw, err := os.OpenFile(temp7z,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			panic(err)
		}
		_, err = fw.Write(z7Data)
		_ = fw.Close()
		if err != nil {
			panic(err)
		}
	}
}

type Z7Err uint8

// CPP\7zip\UI\Common\ExitCode.h
// 注释的几个错误,也没有在7-zip.chm中说明
const (
	KSuccess    Z7Err = 0
	KWarning    Z7Err = 1
	KFatalError Z7Err = 2

	// KCRCError      Z7Err = 3 // A CRC error occurred when unpacking
	// KLockedArchive Z7Err = 4 // Attempt to modify an archive previously locked
	// KWriteError    Z7Err = 5 // Write to disk error
	// KOpenError     Z7Err = 6 // Open file error

	KUserError   Z7Err = 7
	KMemoryError Z7Err = 8

	// KCreateFileError Z7Err = 9 // Create file error

	KUserBreak Z7Err = 255
)

func (e Z7Err) Error() string {
	switch e {
	case KSuccess:
		return "无错误"
	case KWarning:
		// 例如一个或多个文件被其它应用程序锁定，无法添加到压缩档案中
		return "警告 (非致命错误)"
	case KFatalError:
		return "致命错误"
	case KUserError:
		return "命令行错误"
	case KMemoryError:
		return "内存不足"
	case KUserBreak:
		return "用户取消操作"
	default:
		return fmt.Sprintf("ExitCode(%d)", uint8(e))
	}
}

// Path7z 确保外部不可更改7z可执行程序路径
func Path7z() string { return z7path }

// Command 直接返回 *exec.Cmd 对象
func Command(arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: z7path, // 已经通过init成功,这里直接返回对象
		Args: append([]string{z7path}, arg...),
	}
}
