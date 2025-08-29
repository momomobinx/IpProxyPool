package ip66

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/momomobinx/IpProxyPool/fetcher"
	"github.com/momomobinx/IpProxyPool/models/ipModel"
	"github.com/momomobinx/IpProxyPool/util"
	logger "github.com/sirupsen/logrus"
)

func Ip66() []*ipModel.IP {
	logger.Info("[66ip] fetch start")
	defer func() {
		if r := recover(); r != nil {
			logger.Warnln("[66ip] fetch error")
		}
		logger.Warnln("[66ip] fetch error")
	}()
	list := make([]*ipModel.IP, 0)

	indexUrl := "http://www.66ip.cn"
	//fetchIndex := fetcher.Fetch(indexUrl)
	//pageNum := fetchIndex.Find("#PageList > a:nth-child(12)").Text()
	//num, _ := strconv.Atoi(pageNum)
	//fmt.Println("=====")
	//fmt.Println(num)
	for i := 1; i <= 100; i++ {
		url := fmt.Sprintf("%s/%d.html", indexUrl, i)
		fetch := fetcher.Fetch(url)
		fetch.Find("table > tbody").Each(func(i int, selection *goquery.Selection) {
			selection.Find("tr").NextAll().Each(func(i int, selection *goquery.Selection) {
				proxyIp := strings.TrimSpace(selection.Find("td:nth-child(1)").Text())
				proxyPort := strings.TrimSpace(selection.Find("td:nth-child(2)").Text())
				proxyLocation := strings.TrimSpace(selection.Find("td:nth-child(3)").Text())

				ip := new(ipModel.IP)
				ip.ProxyHost = proxyIp
				ip.ProxyPort, _ = strconv.Atoi(proxyPort)
				ip.ProxyType = "http"
				ip.ProxyLocation = proxyLocation
				ip.ProxySpeed = 100
				ip.ProxySource = "http://www.66ip.cn"
				ip.CreateTime = util.FormatDateTime()
				ip.UpdateTime = util.FormatDateTime()
				list = append(list, ip)
			})
		})
	}
	logger.Info("[66ip] fetch done")
	return list
}
