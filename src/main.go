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
	recursionCodes := flag.String("rc", "200,301,302,403,404", "HTTP status codes to recurse on (or 0 to disable)")
	recursionDepth := flag.Int("rd", 2, "Maximum recursion depth")
	flag.Parse()

	if *domain == "" || *wordlist == "" {
		fmt.Printf("%sUsage: ultifuzz -d <domain> -w <wordlist> [-t <threads>] [-ua <user-agent>] [-rc <recursion-codes>] [-rd <recursion-depth>]%s\n", colorCyan, colorReset)
		os.Exit(1)
	}

	fmt.Printf("%s[*] Starting UltiFuzz for %s%s\n", colorBlue, *domain, colorReset)

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

	// 3. Passive Gathering - Wayback, Gau & SQLmap Params
	fmt.Printf("%s[+] Harvesting URLs & Variables for SQLmap...%s\n", colorGreen, colorReset)
	waybackFile := filepath.Join(paramsDir, "wayback.txt")
	gauFile := filepath.Join(paramsDir, "gau.txt")
	sqlmapFile := filepath.Join(paramsDir, "urls_for_sqlmap.txt")
	
	runCommandBash(fmt.Sprintf("waybackurls %s > %s", *domain, waybackFile))
	runCommandBash(fmt.Sprintf("gau %s > %s", *domain, gauFile))
	runCommandBash(fmt.Sprintf("cat %s %s | grep '=' | qsreplace 'test' | sort -u > %s", waybackFile, gauFile, sqlmapFile))

	// 4. Active Fuzzing - ffuf
	fmt.Printf("%s[+] Running Active Fuzzing (ffuf)...%s\n", colorGreen, colorReset)
	ffufOutput := filepath.Join(dirsDir, "ffuf.json")
	
	ffufArgs := []string{
		"-u", fmt.Sprintf("https://%s/FUZZ", *domain),
		"-w", *wordlist,
		"-t", fmt.Sprint(*threads),
		"-H", fmt.Sprintf("User-Agent: %s", *userAgent),
		"-o", ffufOutput,
		"-of", "json",
	}

	// Check if recursion is disabled
	if *recursionCodes == "0" {
		fmt.Printf("%s[*] Recursion disabled. Performing basic fuzzing...%s\n", colorYellow, colorReset)
		ffufArgs = append(ffufArgs, "-mc", "200,301,302,307,401,405") // Only match "open" codes
	} else {
		ffufArgs = append(ffufArgs, "-recursion", "-recursion-strategy", "greedy", "-recursion-depth", fmt.Sprint(*recursionDepth), "-mc", *recursionCodes)
	}
	
	cmd := exec.Command("ffuf", ffufArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	processFfufResults(ffufOutput, filepath.Join(dirsDir, "valid_dirs.txt"))

	fmt.Printf("%s[*] Done! Results: %s%s\n", colorBlue, baseDir, colorReset)
}

func processFfufResults(jsonFile, txtFile string) {
	fmt.Printf("%s[*] Cleaning up results...%s\n", colorBlue, colorReset)
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return
	}

	var ffufOut FfufOutput
	json.Unmarshal(data, &ffufOut)

	file, _ := os.Create(txtFile)
	defer file.Close()

	count := 0
	for _, res := range ffufOut.Results {
		// Only save if status is NOT 404 or 403 (standard success/redirects)
		if res.Status != 404 && res.Status != 403 {
			cleanURL := res.URL
			cleanURL = strings.TrimPrefix(cleanURL, "https://")
			cleanURL = strings.TrimPrefix(cleanURL, "http://")
			fmt.Fprintf(file, "%s %d\n", cleanURL, res.Status)
			count++
		}
	}
	fmt.Printf("%s[!] Found %d valid directories/files. Saved to %s%s\n", colorYellow, count, txtFile, colorReset)
}

func runCommandBash(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Run()
}
