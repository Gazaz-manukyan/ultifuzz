#!/bin/bash

# Цвета для красивого вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}[*] Starting UltiFuzz Global Installer...${NC}"

# Проверка Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}[-] Error: Go is not installed! Please install Go and try again.${NC}"
    exit 1
fi

# Установка зависимостей в ~/go/bin
echo -e "${GREEN}[+] Installing core dependencies (ffuf, subfinder, dnsx, gau, etc.)...${NC}"
go install github.com/ffuf/ffuf/v2@latest
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
go install -v github.com/projectdiscovery/dnsx/cmd/dnsx@latest
go install github.com/lc/gau/v2/cmd/gau@latest
go install github.com/tomnomnom/waybackurls@latest
go install github.com/tomnomnom/assetfinder@latest
go install github.com/tomnomnom/qsreplace@latest

# Сборка и установка самого UltiFuzz
echo -e "${GREEN}[+] Building and installing UltiFuzz globally...${NC}"
cd src
go build -o ultifuzz main.go

# Пытаемся переместить в /usr/local/bin (требуется sudo)
if [ -w /usr/local/bin ]; then
    mv ultifuzz /usr/local/bin/
else
    echo -e "${BLUE}[*] Requesting sudo to install to /usr/local/bin...${NC}"
    sudo mv ultifuzz /usr/local/bin/
fi

cd ..

echo -e "${GREEN}[+] Installation complete!${NC}"
echo -e "${BLUE}[*] You can now run the tool from anywhere by typing: ${GREEN}ultifuzz${NC}"

# Напоминание про PATH для инструментов Go
if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
    echo -e "${RED}[!] Warning: ~/go/bin is not in your PATH.${NC}"
    echo -e "Add this to your .bashrc or .zshrc:"
    echo -e "${BLUE}export PATH=\$PATH:\$(go env GOPATH)/bin${NC}"
fi
