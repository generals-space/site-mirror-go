package crawler

import "regexp"

// SpecialCharsMap 查询参数中的特殊字符
var SpecialCharsMap = map[string]string{
	"\\": "xg",
	":":  "mh",
	"*":  "xh",
	"?":  "wh",
	"<":  "xy",
	">":  "dy",
	"|":  "sx",
	" ":  "kg",
}

// 以如下格式结尾的路径才是可以直接以静态路径访问的.
// 其他如.php, .jsp, .asp等如果nginx中没有相应的处理方法, 无法直接展示.
var htmlURLPatternStr = `(\.(html)|(htm)|(xhtml)|(xml))$`
var htmlURLPattern = regexp.MustCompile(htmlURLPatternStr)

var imagePatternStr = `\.((jpg)|(png)|(bmp)|(jpeg)|(gif)|(webp))$`
var imagePattern = regexp.MustCompile(imagePatternStr)

var fontPatternStr = `\.((ttf)|(woff)|(woff2)|(otf)|(eot))$`
var fontPattern = regexp.MustCompile(fontPatternStr)

// charsetPatternInDOMStr meta[http-equiv]元素, content属性中charset截取的正则模式.
// 如<meta http-equiv="content-type" content="text/html; charset=utf-8">
var charsetPatternInDOMStr = `charset\s*=\s*(\S*)\s*;?`

// charsetPattern 普通的MatchString可直接接受模式字符串, 无需Compile,
// 但是只能作为判断是否匹配, 无法从中获取其他信息.
var charsetPattern = regexp.MustCompile(charsetPatternInDOMStr)

var cssAssetURLPatternStr = `url\(\'(.*?)\'\)|url\(\"(.*?)\"\)|url\((.*?)\)`
var cssAssetURLPattern = regexp.MustCompile(cssAssetURLPatternStr)

var emptyLinkPatternStr = `(^data:)|(^mailto:)|(about:blank)|(javascript:)`
var emptyLinkPattern = regexp.MustCompile(emptyLinkPatternStr)
