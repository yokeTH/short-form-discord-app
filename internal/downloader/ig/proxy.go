package ig

import (
	"bufio"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	ProxyListURL    = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=display_proxies&proxy_format=protocolipport&format=text"
	CheckTimeout    = 5 * time.Second
	CheckTarget     = "https://www.instagram.com"
	RefreshInterval = 15 * time.Minute
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

	go pm.lifecycle()
}

func (pm *ProxyManager) GetClient() *http.Client {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for len(pm.proxies) == 0 {
		log.Warn().Msg("No proxies available, waiting for refresh...")
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

func (pm *ProxyManager) lifecycle() {
	log.Info().Msg("Starting initial proxy fetch...")
	pm.fetchAndScan(true)

	ticker := time.NewTicker(RefreshInterval)
	for range ticker.C {
		log.Info().Msg("Running scheduled proxy refresh (15m)...")
		pm.fetchAndScan(false)
	}
}

func (pm *ProxyManager) fetchAndScan(isStartup bool) {
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

	validChan := make(chan string, len(rawProxies))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 50)

	for _, p := range rawProxies {
		wg.Add(1)
		go func(proxyAddr string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if checkProxy(proxyAddr) {
				validChan <- proxyAddr
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(validChan)
	}()

	if isStartup {
		for p := range validChan {
			pm.addOne(p)
		}
	} else {
		var newProxies []string
		for p := range validChan {
			newProxies = append(newProxies, p)
		}

		if len(newProxies) > 0 {
			pm.replaceAll(newProxies)
		} else {
			log.Warn().Msg("Refresh found 0 valid proxies. Keeping old list.")
		}
	}
}

func (pm *ProxyManager) addOne(p string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.proxies = append(pm.proxies, p)
	pm.cond.Signal()
}

func (pm *ProxyManager) replaceAll(newList []string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	prevCount := len(pm.proxies)
	pm.proxies = newList
	pm.index = 0

	log.Info().Int("old_count", prevCount).Int("new_count", len(newList)).Msg("Proxy list refreshed")

	pm.cond.Broadcast()
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
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   CheckTimeout,
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
