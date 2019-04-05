package crawler

import (
	"net/url"
	"path"
	"strings"
)

// TransToLocalLink ...
// @return: localLink 本地链接, 用于写入本地html文档中的link/script/img/a等标签的链接属性, 格式为以斜线/起始的根路径.
func (crawler *Crawler) TransToLocalLink(fullURL string, urlType int) (localLink string, err error) {
	// 对于域名为host的url, 资源存放目录为output根目录, 而不是域名文件夹. 默认不设置主host
	urlObj, err := url.Parse(fullURL)
	if err != nil {
		logger.Errorf("解析URL出错: %s", err.Error())
		return
	}
	originHost := urlObj.Host
	originPath := urlObj.Path

	localLink = originPath
	if urlType == PageURL {
		localLink = crawler.transToLocalLinkForPage(urlObj)
	} else {
		localLink = crawler.transToLocalLinkForAsset(urlObj)
	}

	// 如果该url就是当前站点域名下的，那么无需新建域名目录存放.
	// 如果是其他站点的(需要事先开启允许下载其他站点静态文件的配置),
	// 则要将资源存放在以站点域名为名的目录下, 路径中仍然需要保留域名部分.
	if originHost != crawler.MainSite {
		host := originHost
		// 有时originHost中可能包含端口, 冒号需要转义.
		host = strings.Replace(host, ":", SpecialCharsMap[":"], -1)
		localLink = "/" + host + localLink
	}
	/*
		// url中可能包含中文(不只是query中), 需要解码.
		localLink, err = url.QueryUnescape(localLink)
		if err != nil {
			logger.Errorf("解码URL出错: localLink: %s, %s", localLink, err.Error())
			return
		}
	*/
	return
}

func (crawler *Crawler) transToLocalLinkForPage(urlObj *url.URL) (localLink string) {
	originPath := urlObj.Path
	originQuery := urlObj.RawQuery

	localLink = originPath

	// 如果path为空
	if localLink == "" {
		localLink = "index.html"
	}
	// 如果path以/结尾
	boolean := strings.HasSuffix(localLink, "/")
	if boolean {
		localLink += "index.html"
	}

	// 替换query参数中的特殊字符
	if originQuery != "" {
		queryStr := originQuery
		for key, val := range SpecialCharsMap {
			queryStr = strings.Replace(queryStr, key, val, -1)
		}
		localLink = localLink + SpecialCharsMap["?"] + queryStr
	}

	// 如果是不支持的页面后缀, 如.php, .jsp, .asp等
	// 注意此时localLink可能是拼接过query的字符串.
	if !htmlURLPattern.MatchString(localLink) {
		localLink += ".html"
	}

	return
}

func (crawler *Crawler) transToLocalLinkForAsset(urlObj *url.URL) (localLink string) {
	originPath := urlObj.Path
	originQuery := urlObj.RawQuery

	localLink = originPath

	// 如果path为空
	if localLink == "" {
		localLink = "index"
	}
	// 如果path以/结尾
	boolean := strings.HasSuffix(localLink, "/")
	if boolean {
		localLink += "index"
	}

	// 替换query参数中的特殊字符
	if originQuery != "" {
		queryStr := originQuery
		for key, val := range SpecialCharsMap {
			queryStr = strings.Replace(queryStr, key, val, -1)
		}
		localLink = localLink + SpecialCharsMap["?"] + queryStr
	}

	return
}

// TransToLocalPath ...
// @return: 返回本地路径与文件名称, 用于写入本地文件
func (crawler *Crawler) TransToLocalPath(fullURL string, urlType int) (fileDir string, fileName string, err error) {
	localLink, err := crawler.TransToLocalLink(fullURL, urlType)

	// 如果是站外资源, local_link可能为/www.xxx.com/static/x.jpg,
	// 但我们需要的存储目录是相对路径, 所以需要事先将链接起始的斜线/移除, 作为相对路径.
	if strings.HasPrefix(localLink, "/") {
		localLink = localLink[1:]
	}

	fileDir = path.Dir(localLink)
	fileName = path.Base(localLink)
	return
}
