package ig

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	ProxyListURL = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=display_proxies&proxy_format=protocolipport&format=text"
	CheckTimeout = 1 * time.Minute
	CheckTarget  = "https://www.instagram.com"
)

type ProxyManager struct {
	proxies []string
	mu      sync.Mutex
	cond    *sync.Cond
	index   int
}

var GlobalProxyManager *ProxyManager

func InitializeProxies() {
	pm := &ProxyManager{}
	pm.cond = sync.NewCond(&pm.mu)
	GlobalProxyManager = pm

	go pm.fetchAndValidate()
}

func (pm *ProxyManager) GetClient() *http.Client {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for len(pm.proxies) == 0 {
		log.Warn().Msg("Waiting for a valid proxy to become available...")
		pm.cond.Wait()
	}

	proxyStr := pm.proxies[pm.index]
	pm.index = (pm.index + 1) % len(pm.proxies)

	if !containsProtocol(proxyStr) {
		proxyStr = "http://" + proxyStr
	}

	proxyURL, _ := url.Parse(proxyStr)

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 15 * time.Second,
	}
}

func (pm *ProxyManager) addProxy(p string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.proxies = append(pm.proxies, p)
	log.Info().Str("proxy", p).Int("total_valid", len(pm.proxies)).Msg("New valid proxy added")

	pm.cond.Signal()
}

func (pm *ProxyManager) fetchAndValidate() {
	log.Info().Msg("Fetching proxy list in background...")
	resp, err := http.Get(ProxyListURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch proxy list")
		return
	}
	defer resp.Body.Close()

	var rawProxies []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			rawProxies = append(rawProxies, line)
		}
	}
	log.Info().Int("count", len(rawProxies)).Msg("Raw proxy list fetched. Starting validation...")

	sem := make(chan struct{}, 50)

	for _, p := range rawProxies {
		go func(proxyAddr string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			if checkProxy(proxyAddr) {
				pm.addProxy(proxyAddr)
			}
		}(p)
	}
}

func validateProxiesConcurrent(list []string) []string {
	var valid []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 50)

	for _, p := range list {
		wg.Add(1)
		go func(proxyAddr string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if checkProxy(proxyAddr) {
				mu.Lock()
				valid = append(valid, proxyAddr)
				mu.Unlock()
				fmt.Print(".")
			}
		}(p)
	}
	wg.Wait()
	fmt.Println()
	return valid
}

func checkProxy(proxyAddr string) bool {
	if !containsProtocol(proxyAddr) {
		proxyAddr = "http://" + proxyAddr
	}

	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: CheckTimeout,
	}

	resp, err := client.Head(CheckTarget)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}

func containsProtocol(addr string) bool {
	return len(addr) > 7 && (addr[:7] == "http://" || addr[:8] == "https://" || addr[:9] == "socks4://" || addr[:9] == "socks5://")
}
