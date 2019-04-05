# goquery页面编码处理(二)-字符实体

1. [ReplaceWithHtml sees html entity where there is none](https://github.com/PuerkitoBio/goquery/issues/113)

2. [Should not escape &](https://github.com/PuerkitoBio/goquery/issues/109)

3. [Don't change &nbsp; to space in Html()](https://github.com/PuerkitoBio/goquery/issues/28)

4. [w3school HTML 字符实体](http://www.w3school.com.cn/html/html_entities.asp)

接着上一篇的问题继续总结.

上一篇想着用第三方编码处理工具[mahonia](https://github.com/axgle/mahonia)进行编解码操作, 但是`&nbsp;`被解析为□, 有点可惜.

后来仔细一看还是有不少问题:

1. `Html()`方法得到的字符串中的某些符号被输出为字符实体编码. 如`'` -> `&#39;`, `&` -> `&amp;`. 一个`a`标签变成了`<a href="javascript:void(0);" onclick="AddFavorite(&#39;https://www.lewenxiaoshuo.com/&#39;,&#39;乐文,乐文小说网,最好的乐文小说阅读网&#39;)">加入收藏</a>`

2. `&nbsp;`实体编码并没有出现在`Html()`结果中, 而且被转换成了一个不可见字符.

------

首先要解析`Html()`结果中的字符实体, 毕竟这不是编码问题. 于是找到了参考文章1和2, 也是goquery的官方issue. 

但作者说这不是goquery的原因, 而是使用了`/x/net/html`包的原因, `/x/net/html`本身会转义一些字符(这个转义的原则我还不清楚), 所以我选择使用`html.UnescapeString()`方法将这个实体反转回来, 得到我想要的.

------

然后是`&nbsp;`的问题, `Html()`的结果中并没有出现`&nbsp;`, 这一点和上面所说的`'` -> `&#39`的行为不同. 但是`&bnsp;`又没有被转义成一个传统的空格`whitespace`(字符编码`\u0020`), 而是转换成了一个不间断空格`non-breaking space`(字符编码`\u00a0`). 见参考文章3(另外, `non-breaking space`的定义可见参考文章4.).

参考文章3中作者提到可以使用`property_test.go`中的`TestNbsp()`函数检测页面中的`&nbsp;`字符. 我使用了如下语句将其替换.

```go
strings.Replace(output, "\u00a0", "&nbsp;", -1)
```

除了`\u00a0`, 还有`©`版权字符等也需要转换.

可以猜想, goquery在调用`Html()`方法时一定是把页面渲染了一次, 才把页面中不支持GBK的字符渲染成了`©`和`\u00a0`等, 我们需要将其全部转回来, 再写入到文件.

...md花了3天时间.