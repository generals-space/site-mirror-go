package crawler

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseLinkingPages 解析并改写页面中的页面链接, 包括a, iframe等元素
func (crawler *Crawler) ParseLinkingPages(htmlDom *goquery.Document, req *RequestTask) {
	aList := htmlDom.Find("a")
	crawler.parseLinkingPages(aList, req, "href")
}

// parseLinkingPages 遍历选中节点, 解析链接入库, 同时修改节点的链接属性.
func (crawler *Crawler) parseLinkingPages(nodeList *goquery.Selection, req *RequestTask, attrName string) {
	// nodeList.Nodes 对象表示当前选择器中包含的元素
	nodeList.Each(func(i int, nodeItem *goquery.Selection) {
		subURL, exist := nodeItem.Attr(attrName)
		if !exist {
			return
		}

		fullURL, fullURLWithoutFrag := joinURL(req.URL, subURL)

		if !URLFilter(fullURL, PageURL, crawler.MainSite) {
			return
		}
		localLink, err := TransToLocalLink(crawler.MainSite, fullURL, PageURL)
		if err != nil {
			return
		}
		nodeItem.SetAttr(attrName, localLink)

		// 新任务入队列
		req := &RequestTask{
			URL:     fullURLWithoutFrag,
			URLType: PageURL,
			Refer:   req.URL,
			Depth:   req.Depth + 1,
		}
		crawler.EnqueuePage(req)
	})
}

// ParseLinkingAssets 解析并改写页面中的静态资源链接, 包括js, css, img等元素
func (crawler *Crawler) ParseLinkingAssets(htmlDom *goquery.Document, req *RequestTask) {
	linkList := htmlDom.Find("link")
	crawler.parseLinkingAssets(linkList, req, "href")

	scriptList := htmlDom.Find("script")
	crawler.parseLinkingAssets(scriptList, req, "src")

	imgList := htmlDom.Find("img")
	crawler.parseLinkingAssets(imgList, req, "src")
}

func (crawler *Crawler) parseLinkingAssets(nodeList *goquery.Selection, req *RequestTask, attrName string) {
	// nodeList.Nodes 对象表示当前选择器中包含的元素
	nodeList.Each(func(i int, nodeItem *goquery.Selection) {
		subURL, exist := nodeItem.Attr(attrName)
		if !exist {
			return
		}

		fullURL, fullURLWithoutFrag := joinURL(req.URL, subURL)
		if !URLFilter(fullURL, AssetURL, crawler.MainSite) {
			return
		}
		localLink, err := TransToLocalLink(crawler.MainSite, fullURL, AssetURL)
		if err != nil {
			return
		}
		nodeItem.SetAttr(attrName, localLink)

		// 新任务入队列
		req := &RequestTask{
			URL:     fullURLWithoutFrag,
			URLType: AssetURL,
			Refer:   req.URL,
			Depth:   req.Depth + 1,
		}
		crawler.EnqueueAsset(req)
	})
}

// parseCSSFile 解析css文件中的链接, 获取资源并修改其引用路径.
// css中可能包含url属性,或者是background-image属性的引用路径,
// 格式可能为url('./bg.jpg'), url("./bg.jpg"), url(bg.jpg)
// 如下， import_list可能是[('', '', 'bg.jpg'), ('', '', 'logo.png')]
// 元组中前两个空格表示匹配到的都是url(bg.jpg)这种形式的属性
func (crawler *Crawler) parseCSSFile(content []byte, req *RequestTask) (newContent []byte, err error) {
	fileStr := string(content)
	// FindAllStringSubmatch返回值为切片, 是所有匹配到的字符串集合.
	// 其成员也是切片, 此切片类似于FindStringSubmatch()的结果, 表示分组的匹配情况.
	matchedArray := cssAssetURLPattern.FindAllStringSubmatch(fileStr, -1)
	for _, matchedItem := range matchedArray {
		for _, matchedURL := range matchedItem[1:] {
			if matchedURL == "" || emptyLinkPattern.MatchString(matchedURL) {
				continue
			}
			fullURL, fullURLWithoutFrag := joinURL(req.URL, matchedURL)
			if !URLFilter(fullURL, AssetURL, crawler.MainSite) {
				return
			}
			localLink, err := TransToLocalLink(crawler.MainSite, fullURL, AssetURL)
			if err != nil {
				continue
			}
			fileStr = strings.Replace(fileStr, matchedURL, localLink, -1)
			// 新任务入队列
			req := &RequestTask{
				URL:     fullURLWithoutFrag,
				URLType: AssetURL,
				Refer:   req.URL,
				Depth:   req.Depth + 1,
			}
			crawler.EnqueueAsset(req)
		}
	}
	newContent = []byte(fileStr)
	return
}
