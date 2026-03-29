package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

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
	
	// Use dnsx to verify active subdomains
	activeSubFile := filepath.Join(subdomainsDir, "active_subdomains.txt")
	runCommandBash(fmt.Sprintf("cat %s | sort -u | dnsx -silent > %s", rawSubFile, activeSubFile))
	fmt.Printf("%s[!] Active subdomains saved to: %s%s\n", colorYellow, activeSubFile, colorReset)

	// 3. Passive Gathering - Wayback & Params
	fmt.Printf("%s[+] Running Wayback & Archive Gathering (waybackurls, gau)...%s\n", colorGreen, colorReset)
	waybackFile := filepath.Join(paramsDir, "wayback.txt")
	gauFile := filepath.Join(paramsDir, "gau.txt")
	
	runCommandBash(fmt.Sprintf("waybackurls %s > %s", *domain, waybackFile))
	runCommandBash(fmt.Sprintf("gau %s > %s", *domain, gauFile))

	// Extracting params and combining
	paramsFile := filepath.Join(paramsDir, "params.txt")
	combinedUrls := filepath.Join(paramsDir, "combined_urls.txt")
	runCommandBash(fmt.Sprintf("cat %s %s | sort -u > %s", waybackFile, gauFile, combinedUrls))
	runCommandBash(fmt.Sprintf("grep '?' %s | cut -d '?' -f 2 | tr '&' '\\n' | sort -u > %s", combinedUrls, paramsFile))
	fmt.Printf("%s[!] Extracted parameters saved to: %s%s\n", colorYellow, paramsFile, colorReset)

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
	
	// For ffuf we want to see the output
	cmd := exec.Command("ffuf", ffufArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[-] ffuf execution finished with some issues (or stopped by user): %v%s\n", colorYellow, err, colorReset)
	}

	fmt.Printf("%s[*] Done! Results saved in %s%s\n", colorBlue, baseDir, colorReset)
}

func runCommandBash(command string) {
	cmd := exec.Command("bash", "-c", command)
	// We don't pipe stdout for passive tools to keep it clean, 
	// but you could pipe it to a log file if needed.
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[-] Warning: command failed: %s (%v)%s\n", colorYellow, command, err, colorReset)
	}
}
