package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

func main() {
	// Fetch HaGeZi Ultimate list
	url := "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/ultimate.txt"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	domains := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domains = append(domains, line)
	}

	domainToIPs := make(map[string][]string)
	var mu sync.Mutex // To safely write to the map
	var wg sync.WaitGroup

	// Limit the number of concurrent Goroutines
	concurrency := 20
	sem := make(chan struct{}, concurrency)

	count := 0
	for _, domain := range domains {
		wg.Add(1)
		sem <- struct{}{} // acquire slot

		go func(d string) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			ips, err := net.LookupIP(d)
			if err != nil {
				return // skip domains that fail to resolve
			}

			ipStrs := []string{}
			for _, ip := range ips {
				ipStrs = append(ipStrs, ip.String())
			}

			mu.Lock()
			domainToIPs[d] = ipStrs
			count++
			fmt.Printf("[%d] %s -> %s\n", count, d, strings.Join(ipStrs, ", "))
			mu.Unlock()
		}(domain)
	}

	wg.Wait()
	fmt.Printf("\nResolved %d domains in total\n", count)
}
