package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// FfufResult represents a partial structure of ffuf output
type FfufOutput struct {
	Results []FfufResult `json:"results"`
}

type FfufResult struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
}

func main() {
	domain := flag.String("d", "", "Target domain (e.g., example.com)")
	wordlist := flag.String("w", "", "Path to wordlist")
	threads := flag.Int("t", 40, "Number of threads")
	userAgent := flag.String("ua", "UltiFuzz/1.0", "User-Agent string")
	recursionCodes := flag.String("rc", "200,301,302,403,404", "HTTP status codes to recurse on")
	recursionDepth := flag.Int("rd", 2, "Maximum recursion depth")
	flag.Parse()

	if *domain == "" || *wordlist == "" {
		fmt.Printf("%sUsage: ultifuzz -d <domain> -w <wordlist> [-t <threads>] [-ua <user-agent>] [-rc <recursion-codes>] [-rd <recursion-depth>]%s\n", colorCyan, colorReset)
		os.Exit(1)
	}

	fmt.Printf("%s[*] Starting UltiFuzz for %s%s\n", colorBlue, *domain, colorReset)

	// 1. Create directory structure
	baseDir := filepath.Join("output", *domain)
	dirsDir := filepath.Join(baseDir, "dirs")
	subdomainsDir := filepath.Join(baseDir, "subdomains")
	paramsDir := filepath.Join(baseDir, "params")

	for _, dir := range []string{dirsDir, subdomainsDir, paramsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("%s[-] Error creating directory %s: %v%s", colorRed, dir, err, colorReset)
		}
	}

	// 2. Passive Gathering - Subdomains
	fmt.Printf("%s[+] Running Subdomain Gathering...%s\n", colorGreen, colorReset)
	rawSubFile := filepath.Join(subdomainsDir, "raw_subdomains.txt")
	runCommandBash(fmt.Sprintf("subfinder -d %s -silent > %s", *domain, rawSubFile))
	runCommandBash(fmt.Sprintf("assetfinder --subs-only %s >> %s", *domain, rawSubFile))
	
	activeSubFile := filepath.Join(subdomainsDir, "active_subdomains.txt")
	runCommandBash(fmt.Sprintf("cat %s | sort -u | dnsx -silent > %s", rawSubFile, activeSubFile))
	fmt.Printf("%s[!] Active subdomains saved to: %s%s\n", colorYellow, activeSubFile, colorReset)

	// 3. Passive Gathering - Wayback, Gau & SQLmap Params
	fmt.Printf("%s[+] Harvesting URLs & Variables for SQLmap...%s\n", colorGreen, colorReset)
	waybackFile := filepath.Join(paramsDir, "wayback.txt")
	gauFile := filepath.Join(paramsDir, "gau.txt")
	sqlmapFile := filepath.Join(paramsDir, "urls_for_sqlmap.txt")
	
	runCommandBash(fmt.Sprintf("waybackurls %s > %s", *domain, waybackFile))
	runCommandBash(fmt.Sprintf("gau %s > %s", *domain, gauFile))

	runCommandBash(fmt.Sprintf("cat %s %s | grep '=' | qsreplace 'test' | sort -u > %s", waybackFile, gauFile, sqlmapFile))
	fmt.Printf("%s[!] SQLmap-ready URLs saved to: %s%s\n", colorYellow, sqlmapFile, colorReset)

	// 4. Active Fuzzing - ffuf
	fmt.Printf("%s[+] Running Active Fuzzing (ffuf)...%s\n", colorGreen, colorReset)
	ffufOutput := filepath.Join(dirsDir, "ffuf.json")
	
	ffufArgs := []string{
		"-u", fmt.Sprintf("https://%s/FUZZ", *domain),
		"-w", *wordlist,
		"-t", fmt.Sprint(*threads),
		"-H", fmt.Sprintf("User-Agent: %s", *userAgent),
		"-recursion",
		"-recursion-strategy", "greedy",
		"-recursion-depth", fmt.Sprint(*recursionDepth),
		"-mc", *recursionCodes,
		"-o", ffufOutput,
		"-of", "json",
	}
	
	cmd := exec.Command("ffuf", ffufArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[-] ffuf execution finished with some issues (or stopped by user): %v%s\n", colorYellow, err, colorReset)
	}

	// 5. Finalize - Processing results for clean output
	processFfufResults(ffufOutput, filepath.Join(dirsDir, "valid_dirs.txt"))

	fmt.Printf("%s[*] Done! Final results saved in %s%s\n", colorBlue, baseDir, colorReset)
}

func processFfufResults(jsonFile, txtFile string) {
	fmt.Printf("%s[*] Cleaning up results...%s\n", colorBlue, colorReset)
	
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("%s[-] Could not read ffuf results: %v%s\n", colorRed, err, colorReset)
		return
	}

	var ffufOut FfufOutput
	if err := json.Unmarshal(data, &ffufOut); err != nil {
		fmt.Printf("%s[-] Error parsing ffuf JSON: %v%s\n", colorRed, err, colorReset)
		return
	}

	file, err := os.Create(txtFile)
	if err != nil {
		fmt.Printf("%s[-] Error creating valid_dirs.txt: %v%s\n", colorRed, err, colorReset)
		return
	}
	defer file.Close()

	count := 0
	for _, res := range ffufOut.Results {
		// Only save if status is NOT 404 or 403
		if res.Status != 404 && res.Status != 403 {
			// Clean URL from http:// or https://
			cleanURL := res.URL
			cleanURL = strings.TrimPrefix(cleanURL, "https://")
			cleanURL = strings.TrimPrefix(cleanURL, "http://")
			
			// Format: domen/dir status_code
			fmt.Fprintf(file, "%s %d\n", cleanURL, res.Status)
			count++
		}
	}
	fmt.Printf("%s[!] Found %d valid directories/files. Saved to %s%s\n", colorYellow, count, txtFile, colorReset)
}

func runCommandBash(command string) {
	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[-] Warning: command failed: %s (%v)%s\n", colorYellow, command, err, colorReset)
	}
}
