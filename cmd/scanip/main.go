package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/malfunkt/iprange"
)

// Configurable Defaults
const (
	defaultThreads = 500
	defaultTimeout = 280 * time.Millisecond
	defaultPort    = "80"
	resultsFile    = "scanip.results.txt"
)

var (
	verbose bool
	wg      sync.WaitGroup
	mu      sync.Mutex
	scanned = 0
	total   = 0
)

// User agents list (randomized selection)
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Firefox/89.0",
}

// Selects a random user agent
func randomUserAgent() string {
	var n uint32
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return userAgents[n%uint32(len(userAgents))]
}

// Fetches the title of the page
func getTitle(body string) string {
	re := regexp.MustCompile(`<title>(.*?)</title>`)
	match := re.FindStringSubmatch(body)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return "No Title"
}

// Fetches the website content
func fetchWebsite(ip, port string) {
	defer wg.Done()
	url := fmt.Sprintf("http://%s:%s", ip, port)

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", randomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read up to 100KB of response
	body, err := io.ReadAll(io.LimitReader(resp.Body, 102400))
	if err != nil {
		return
	}

	title := getTitle(string(body))
	sizeKB := len(body) / 1024

	// Format result
	result := fmt.Sprintf("%s,%s,%s,%dKB", ip, port, title, sizeKB)

	// Append result to file safely
	mu.Lock()
	appendToFile(resultsFile, result)
	scanned++
	mu.Unlock()

	// Display in green
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("%s/%s | %s:%s | %s | %dKB\n", green(scanned), blue(total), green(ip), blue(port), green(title), sizeKB)
}

// Appends a line to the file safely
func appendToFile(filename, line string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, line)
}

// Expands CIDR/IP ranges to a list of IPs
func expandIPs(input string) []string {
	expanded := []string{}
	ranges := strings.Split(input, ",")

	for _, r := range ranges {
		r = strings.TrimSpace(r)
		list, err := iprange.ParseList(r)
		if err == nil {
			for _, ip := range list.Expand() {
				expanded = append(expanded, ip.String()) // Convert net.IP to string
			}
		}

	}

	return expanded
}

// Prints usage instructions
func printHelp() {
	fmt.Println("Usage: scanip [OPTIONS]")
	fmt.Println("\nOptions:")
	fmt.Println("  -h          Show this help message and exit")
	fmt.Println("  -v          Enable verbose output\n")
	fmt.Println("Example:")
	fmt.Println("  ./scanip")
	fmt.Println("  (Prompts for CIDR, threads, and port)")
}

// Main function
func main() {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-h" {
			printHelp()
			return
		} else if arg == "-v" {
			verbose = true
		}
	}

	// Prompt for CIDR notation or IP range
	fmt.Print("Enter CIDR(s) or IP range (comma-separated): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	cidrInput := scanner.Text()
	if cidrInput == "" {
		fmt.Println("No input provided. Exiting.")
		return
	}

	ips := expandIPs(cidrInput)
	total = len(ips)

	if total == 0 {
		fmt.Println("No valid IPs found. Exiting.")
		return
	}

	// Prompt for number of threads
	fmt.Printf("Enter number of threads (default %d): ", defaultThreads)
	scanner.Scan()
	threadsInput := scanner.Text()
	threads := defaultThreads
	if threadsInput != "" {
		fmt.Sscanf(threadsInput, "%d", &threads)
	}

	// Prompt for port
	fmt.Printf("Enter website port (default %s): ", defaultPort)
	scanner.Scan()
	portInput := scanner.Text()
	port := defaultPort
	if portInput != "" {
		port = portInput
	}

	// Start scanning with worker pool
	sem := make(chan struct{}, threads)

	for _, ip := range ips {
		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer func() { <-sem }()
			fetchWebsite(ip, port)
		}(ip)
	}

	wg.Wait()
	fmt.Println("Scan completed.")
}
