package crawler

import (
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// ParseLinkingPages 解析并改写页面中的页面链接, 包括a, iframe等元素
func (crawler *Crawler) ParseLinkingPages(htmlDom *goquery.Document, req *RequestTask) {
	aList := htmlDom.Find("a")
	crawler.parseLinkingPages(aList, req, "href")
}

// parseLinkingPages 遍历选中节点, 解析链接入库, 同时修改节点的链接属性.
func (crawler *Crawler) parseLinkingPages(nodeList *goquery.Selection, req *RequestTask, attrName string) {
	logger.Infof("length :%d", nodeList.Length())
	// nodeList.Nodes 对象表示当前选择器中包含的元素
	nodeList.Each(func(i int, nodeItem *goquery.Selection) {
		subURL, exist := nodeItem.Attr(attrName)
		if !exist {
			return
		}

		baseURLObj, _ := url.Parse(req.URL)
		subURLObj, _ := url.Parse(subURL)
		fullURLObj := baseURLObj.ResolveReference(subURLObj)
		fullURL := fullURLObj.String()
		fullURLObj.Fragment = ""
		fullURLWithoutFrag := fullURLObj.String()

		if !URLFilter(fullURL, PageURL, crawler.MainSite) {
			return
		}
		localLink, err := crawler.TransToLocalLink(fullURL, PageURL)
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

}
