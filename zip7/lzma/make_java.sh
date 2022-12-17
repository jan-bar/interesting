# 因为要用到bash的find命令而不是window的find,所以及时在window也用bash执行
# bash -x make_java.sh C\lzma2201\Java
old=$(PWD)
cd $1
echo -e 'Manifest-Version: 1.0\nMain-Class: SevenZip.LzmaAlone' > manifest
/bin/find . -name "*.java" > java.txt
javac @java.txt # 编译所有java文件为class文件
/bin/find . -name "*.class" > class.txt
jar -cvfm $old/lzma.jar manifest @class.txt # 打包所有class文件
cd $old
