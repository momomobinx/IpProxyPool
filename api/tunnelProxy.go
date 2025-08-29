package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/momomobinx/IpProxyPool/middleware/config"
	"github.com/momomobinx/IpProxyPool/middleware/storage"
	"github.com/momomobinx/IpProxyPool/models/ipModel"
)

var httpIp string
var httpsIp string
var socket5Ip string

func HttpSRunTunnelProxyServer(conf *config.Tunnel) {
	httpsIp = getHttpsIp()
	httpIp = getHttpIp()
	var httpCurCount int
	var httpsCurCount int
	httpCurCount = 0
	httpsCurCount = 0

	log.Println("HTTP 隧道代理启动 - 监听IP端口 -> ", conf.Ip+":"+conf.HttpTunnelPort)

	server := &http.Server{
		Addr:      conf.Ip + ":" + conf.HttpTunnelPort,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method == http.MethodConnect {
				if httpsCurCount < conf.UseCount {
					httpsCurCount++
				} else {
					httpsCurCount = 0
					httpsIp = getHttpsIp()
				}
				log.Printf("隧道代理 | HTTPS 请求：%s 使用代理: %s", r.URL.String(), httpsIp)
				destConn, err := net.DialTimeout("tcp", httpsIp, 20*time.Second)
				if err != nil {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
					return
				}
				destConn.SetReadDeadline(time.Now().Add(20 * time.Second))
				var req []byte
				req = MergeArray([]byte(fmt.Sprintf("%s %s %s%s", r.Method, r.Host, r.Proto, []byte{13, 10})), []byte(fmt.Sprintf("Host: %s%s", r.Host, []byte{13, 10})))
				for k, v := range r.Header {
					req = MergeArray(req, []byte(fmt.Sprintf(
						"%s: %s%s", k, v[0], []byte{13, 10})))
				}
				req = MergeArray(req, []byte{13, 10})
				io.ReadAll(r.Body)
				all, err := io.ReadAll(r.Body)
				if err == nil {
					req = MergeArray(req, all)
				}
				destConn.Write(req)
				w.WriteHeader(http.StatusOK)
				hijacker, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "not supported", http.StatusInternalServerError)
					return
				}
				clientConn, _, err := hijacker.Hijack()
				if err != nil {
					return
				}
				clientConn.SetReadDeadline(time.Now().Add(20 * time.Second))
				destConn.Read(make([]byte, 1024)) //先读取一次
				go io.Copy(destConn, clientConn)
				go io.Copy(clientConn, destConn)

			} else {
				if httpCurCount < conf.UseCount {
					httpCurCount++
				} else {
					httpCurCount = 0
					httpIp = getHttpIp()
				}

				log.Printf("隧道代理 | HTTP 请求：%s 使用代理: %s", r.URL.String(), httpIp)
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				//配置代理
				proxyUrl, parseErr := url.Parse("http://" + httpIp)
				if parseErr != nil {
					return
				}
				tr.Proxy = http.ProxyURL(proxyUrl)
				client := &http.Client{Timeout: 20 * time.Second, Transport: tr}
				request, err := http.NewRequest(r.Method, "", r.Body)
				//增加header选项
				request.URL = r.URL
				request.Header = r.Header
				//处理返回结果
				res, err := client.Do(request)
				if err != nil {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
					return
				}
				defer res.Body.Close()

				for k, vv := range res.Header {
					for _, v := range vv {
						w.Header().Add(k, v)
					}
				}
				var bodyBytes []byte
				bodyBytes, _ = io.ReadAll(res.Body)
				res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				w.WriteHeader(res.StatusCode)
				io.Copy(w, res.Body)
				res.Body.Close()

			}
		}),
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func Socket5RunTunnelProxyServer(conf *config.Tunnel) {
	socket5Ip = getSocket5Ip()
	var curCount int
	curCount = 0

	log.Println("SOCKET5 隧道代理启动 - 监听IP端口 -> ", conf.Ip+":"+conf.SocketTunnelPort)
	li, err := net.Listen("tcp", conf.Ip+":"+conf.SocketTunnelPort)
	if err != nil {
		log.Println(err)
	}
	for {
		if curCount < conf.UseCount {
			curCount++
		} else {
			curCount = 0
			socket5Ip = getSocket5Ip()
		}

		clientConn, err := li.Accept()
		if err != nil {
			log.Panic(err)
		}
		go func() {
			log.Printf("隧道代理 | SOCKET5 请求 使用代理: %s", socket5Ip)
			if clientConn == nil {
				return
			}
			defer clientConn.Close()
			destConn, err := net.DialTimeout("tcp", socket5Ip, 30*time.Second)
			if err != nil {
				log.Println(err)
				return
			}
			defer destConn.Close()

			go io.Copy(destConn, clientConn)
			io.Copy(clientConn, destConn)
		}()
	}

}

// MergeArray 合并数组
func MergeArray(dest []byte, src []byte) (result []byte) {
	result = make([]byte, len(dest)+len(src))
	//将第一个数组传入result
	copy(result, dest)
	//将第二个数组接在尾部，也就是 len(dest):
	copy(result[len(dest):], src)
	return
}

func getIp(proxyType string) string {
	ip := ipModel.IP{}
	ip = storage.RandomByProxyType(proxyType)
	addr := ""
	addr = ip.ProxyHost + ":" + strconv.Itoa(ip.ProxyPort)
	return addr
}

func getHttpIp() string {
	return getIp("http")
}

func getHttpsIp() string {
	return getIp("https")
}

func getSocket5Ip() string {
	return getIp("socks")
}
