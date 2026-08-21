#!/usr/bin/env bash
# komutcu kurulum betigi - CachyOS / Arch
set -e

echo ">> Go kurulu mu kontrol ediliyor..."
if ! command -v go >/dev/null 2>&1; then
    echo ">> Go bulunamadi, kuruluyor..."
    sudo pacman -S --needed --noconfirm go
fi

echo ">> Derleniyor..."
go build -o komutcu .

echo ">> /usr/local/bin altina kopyalaniyor..."
sudo install -Dm755 komutcu /usr/local/bin/komutcu

echo ""
echo "Kurulum tamam. Artik her yerden su sekilde calistirabilirsin:"
echo "   komutcu"
echo "   komutcu wifi baglan"
