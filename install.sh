#!/bin/bash
# Furryjan - e621 Content Downloader 
# Installation script for Linux

# Dependencies:
# - go-sqlite3: SQLite3 database driver
# - progressbar/v3: Progress bar for download progress
# - x/crypto: Cryptographic packages
# - x/term: Terminal utilities
# - x/sys: System call wrappers
# - pkg/sftp: SFTP support
# - colorstring: Terminal color support
# - uniseg: Unicode segmentation

set -e

# Change to src directory where go.mod is located
cd "$(dirname "$0")/src"

echo "🔨 Building Furryjan..."
go mod tidy
go build -o furryjan ./cmd/main.go

echo "📦 Installing to /usr/bin/furryjan..."
sudo mv furryjan /usr/bin/furryjan
sudo chmod +x /usr/bin/furryjan

# Install i18n locales to /usr/share/furryjan
echo "📝 Installing translations..."
sudo mkdir -p /usr/share/furryjan/locales
if [ -d "i18n/locales" ]; then
    sudo cp -r i18n/locales/* /usr/share/furryjan/locales/
fi

# Create .desktop file for application menu
echo "🎯 Adding to application menu..."
DESKTOP_FILE="/usr/share/applications/furryjan.desktop"
sudo tee "$DESKTOP_FILE" > /dev/null << 'EOF'
[Desktop Entry]
Version=1.0
Type=Application
Name=Furryjan
Comment=e621 Content Downloader
Exec=/usr/bin/furryjan
Terminal=true
Categories=Utility;Network;FileTransfer;
Icon=package-download
StartupNotify=true
EOF

sudo chmod 644 "$DESKTOP_FILE"

echo "✅ Installation complete!"
echo "Run 'furryjan' to start the application."
echo "You can also find it in your application menu."
