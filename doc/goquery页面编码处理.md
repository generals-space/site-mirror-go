# goquery页面编码处理

1. [goquery - Handle Non-UTF8 html Pages](https://github.com/PuerkitoBio/goquery/wiki/Tips-and-tricks)

2. [goquery 增加GBK支持](https://blog.csdn.net/jrainbow/article/details/52712685)

3. [Golang的字符编码介绍](https://www.cnblogs.com/yinzhengjie/p/7956689.html)

4. [colly 抓取页面乱码问题](https://studygolang.com/topics/6745)

5. [[go]“编码时不支持符号”utf8到sjis转换时出错](https://teratail.com/questions/106106)

6. [Best way to translate UTF-8 to ISO8859-1 in Go](https://stackoverflow.com/questions/47660160/best-way-to-translate-utf-8-to-iso8859-1-in-go)

goquery的`NewDocumentXXX()`函数默认只接收utf-8编码的页面内容, 编码转换操作需要用户自行处理. (查了下, 另一个dom解析工具colly也是这样).

有人在issue中提到`GBK`编码和`CJK(Chinene, Japanese, Korean)`的支持问题, goquery作者回复可见wiki, 即本文参考文章1. 其中提到可以使用[iconv-go](https://github.com/djimenez/iconv-go)包, 这个包实际使用的是C语言中的`iconv`函数, 需要cgo支持, 我在win下无法安装此包, linux下没试过, 放弃了.

考虑到我抓取的目标网页中可能出现的编码有限, 不外乎如下几种

```go
var CharsetMap = map[string]encoding.Encoding{
	"utf-8":   unicode.UTF8,
	"gbk":     simplifiedchinese.GBK,
	"gb2312":  simplifiedchinese.GB18030,
	"gb18030": simplifiedchinese.GB18030,
	"big5":    traditionalchinese.Big5,
}
```

于是我在网上找到参考文章2和3, 尝试了一下`golang.org/x/text`简单的编解码操作.

```go
// DecodeToUTF8 从输入的byte数组中按照指定的字符集解析出对应的utf8格式的内容并返回.
func DecodeToUTF8(input []byte, charset encoding.Encoding) (output []byte, err error) {
	reader := transform.NewReader(bytes.NewReader(input), charset.NewDecoder())
	output, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return
}

// EncodeFromUTF8 将输入的utf-8格式的byte数组中按照指定的字符集编码并返回
func EncodeFromUTF8(input []byte, charset encoding.Encoding) (output []byte, err error) {
	reader := transform.NewReader(bytes.NewReader(input), charset.NewEncoder())
	output, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return
}
```

本来实验时很正常, 编解码和读写文件没有什么问题. 

但是使用goquery的`NewDocument()`读取http响应(已进行utf-8解码)构建dom对象后, 通过`Html()`方法得到页面内容, 我想将页面内容按照页面原始编码写入文件.

然后在调用`EncodeFromUTF8()`进行编码操作时出错, 报`encoding: rune not supported by encoding.`

这个问题是因为某个字符存在于utf-8但是不存在于gbk, 见参考文章5. 经过排查, 发现经过`goquery`处理的, 是页面中的`&nbsp;`字符还原成原编码内容时的地方报此错误.

------

本来想换成colloy的, 但是发现colloy也有这个问题, 不过在参考文章4中找到[mahonia](https://github.com/axgle/mahonia), 试了下, 还好.

不过`&bnsp;`按原编码写到文件中变成了方框...

```
<a href="https://www.lewenxiaoshuo.com" target="_blank" title=""></a>
```

按照参考文章6的提示, `/x/text`包中有一个`ReplaceUnsupported()`方法, 改写`EncodeFromUTF8()`方法为如下

```go
// EncodeFromUTF8 将输入的utf-8格式的byte数组中按照指定的字符集编码并返回
func EncodeFromUTF8(input []byte, charset encoding.Encoding) (output []byte, err error) {
	if charset == unicode.UTF8 {
		output = input
		return
	}
	reader := transform.NewReader(bytes.NewReader(input), encoding.ReplaceUnsupported(charset.NewEncoder()))
	output, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return
}
```

这样会忽略不支持的字符, 也会得到□, 但总归不会出错, 也不用`mahonia`了.