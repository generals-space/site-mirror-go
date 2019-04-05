package crawler

import (
	"bytes"
	"html"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// CharsetMap 字符集映射
var CharsetMap = map[string]encoding.Encoding{
	"utf-8":   unicode.UTF8,
	"gbk":     simplifiedchinese.GBK,
	"gb2312":  simplifiedchinese.GB18030,
	"gb18030": simplifiedchinese.GB18030,
	"big5":    traditionalchinese.Big5,
}

// HTMLCharacterEntitiesMap HTML 字符实体
var HTMLCharacterEntitiesMap = map[string]string{
	"\u00a0": "&nbsp;",
	"©":      "&copy;",
	"®":      "&reg;",
	"™":      "&trade;",
	"￠":      "&cent;",
	"£":      "&pound;",
	"¥":      "&yen;",
	"€":      "&euro;",
	"§":      "&sect;",
}

// ReplaceHTMLCharacterEntities 替换页面中html实体字符, 以免写入文件时遇到不支持的字符
func ReplaceHTMLCharacterEntities(input string, charset encoding.Encoding) (output string) {
	if charset == unicode.UTF8 {
		output = input
		return
	}
	output = html.UnescapeString(input)
	for char, entity := range HTMLCharacterEntitiesMap {
		output = strings.Replace(output, char, entity, -1)
	}
	return
}

// DecodeToUTF8 从输入的byte数组中按照指定的字符集解析出对应的utf8格式的内容并返回.
func DecodeToUTF8(input []byte, charset encoding.Encoding) (output []byte, err error) {
	if charset == unicode.UTF8 {
		output = input
		return
	}
	reader := transform.NewReader(bytes.NewReader(input), charset.NewDecoder())
	output, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return
}

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
