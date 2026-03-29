# UltiFuzz 🚀

**UltiFuzz** is a high-performance, automated reconnaissance and fuzzing orchestrator. It is designed to maximize discovery by combining passive data gathering with aggressive, "greedy" recursive fuzzing.

## 🌟 Why UltiFuzz?

Unlike standard fuzzers that stop at a `404 Not Found` or `403 Forbidden`, **UltiFuzz** can be configured to keep digging. If a directory like `/admin/` returns 404, UltiFuzz will still try to fuzz its children (e.g., `/admin/config.php`) because, in modern web applications, intermediate paths are often hidden while the endpoints remain accessible.

## 🛠 Features

- **Passive Recon:**
  - Fast subdomain discovery via `subfinder` & `assetfinder`.
  - DNS resolution and validation with `dnsx`.
  - Massive URL harvesting from archives using `waybackurls` & `gau`.
  - Automated parameter extraction for further vulnerability research.
- **Aggressive Fuzzing:**
  - Built on top of `ffuf` (Fuzz Faster U Fool).
  - **Greedy Recursion:** Fuzzes deeper even if the parent path returns 403/404.
  - Custom recursion status codes & depth.
- **Smart Organization:**
  - Automatically creates a clean folder structure: `output/<domain>/{subdomains,dirs,params}`.

## 🚀 Installation (Linux/macOS)

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/ultifuzz.git
   cd ultifuzz
   ```

2. Run the global installer:
   ```bash
   chmod +x install.sh
   ./install.sh
   ```

3. **Verify:** You can now run the tool from any directory by typing `ultifuzz`.

## 📖 Usage

```bash
ultifuzz -d target.com -w /path/to/wordlist.txt
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-d` | Target domain | (Required) |
| `-w` | Path to wordlist | (Required) |
| `-t` | Concurrent threads | 40 |
| `-rc`| Status codes to recurse | 200,301,302,403,404 |
| `-rd`| Max recursion depth | 2 |
| `-ua`| Custom User-Agent | UltiFuzz/1.0 |

## 📁 Output Structure

```text
output/
└── example.com/
    ├── subdomains/       # Found and active subdomains
    ├── dirs/             # Ffuf JSON results
    └── params/           # Extracted URLs and query parameters
```

## ⚠️ Requirements

- **Go** (1.19+) installed and in your PATH.
- Make sure `~/go/bin` is in your environment's `$PATH`.
