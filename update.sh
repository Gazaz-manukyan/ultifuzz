#!/bin/bash

# Цвета
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

echo -e "${BLUE}[*] Starting clean update for UltiFuzz...${NC}"

if [ ! -d .git ]; then
    echo -e "${RED}[-] Error: .git directory not found. Please clone the repo again.${NC}"
    exit 1
fi

# 1. Удаляем старый бинарник из системы
echo -e "${BLUE}[*] Removing old binary...${NC}"
if [ -f /usr/local/bin/ultifuzz ]; then
    sudo rm /usr/local/bin/ultifuzz
fi

# 2. Обновляем исходный код из GitHub (Hard Reset)
echo -e "${BLUE}[*] Fetching latest source code...${NC}"
git fetch origin main &> /dev/null
git reset --hard origin/main

# 3. Запускаем чистую установку
echo -e "${BLUE}[*] Reinstalling...${NC}"
chmod +x install.sh
./install.sh

echo -e "${GREEN}[+] UltiFuzz has been updated and reinstalled successfully!${NC}"
