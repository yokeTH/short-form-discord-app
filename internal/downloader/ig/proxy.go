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
	index   int
}

var GlobalProxyManager *ProxyManager

func InitializeProxies() error {
	log.Info().Msg("Fetching proxy list...")
	resp, err := http.Get(ProxyListURL)
	if err != nil {
		return fmt.Errorf("failed to fetch proxy list: %v", err)
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

	log.Info().Int("count", len(rawProxies)).Msg("Proxies fetched. Validating...")

	validProxies := validateProxiesConcurrent(rawProxies)
	if len(validProxies) == 0 {
		return fmt.Errorf("no valid proxies found")
	}

	log.Info().Int("valid_count", len(validProxies)).Msg("Proxy validation complete")

	GlobalProxyManager = &ProxyManager{
		proxies: validProxies,
		index:   0,
	}
	return nil
}

func (pm *ProxyManager) GetClient() *http.Client {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.proxies) == 0 {
		log.Warn().Msg("No proxies available, returning default client")
		return &http.Client{Timeout: 10 * time.Second}
	}

	proxyStr := pm.proxies[pm.index]
	pm.index = (pm.index + 1) % len(pm.proxies)

	if !containsProtocol(proxyStr) {
		proxyStr = "http://" + proxyStr
	}

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		log.Error().Err(err).Str("proxy", proxyStr).Msg("Failed to parse proxy URL")
		return &http.Client{Timeout: 10 * time.Second}
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 15 * time.Second,
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
