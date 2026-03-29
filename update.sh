#!/bin/bash

# Цвета для красивого вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}[*] Checking for updates...${NC}"

# Проверка, находимся ли мы в git-репозитории
if [ ! -d .git ]; then
    echo -e "${RED}[-] Error: This folder is not a git repository. Please clone it from GitHub again.${NC}"
    exit 1
fi

# Получаем изменения
git fetch origin main
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)

if [ "$LOCAL" = "$REMOTE" ]; then
    echo -e "${GREEN}[+] UltiFuzz is already up to date!${NC}"
else
    echo -e "${BLUE}[*] Update found. Pulling latest changes...${NC}"
    git pull origin main
    
    # Пересобираем и устанавливаем
    chmod +x install.sh
    ./install.sh
    
    echo -e "${GREEN}[+] UltiFuzz has been updated to the latest version!${NC}"
fi
