# Linux Installation

## Method 1: Using Simple Installer Scripts

Download and run the installer script directly:

```bash
# Download and install to /usr/local/bin (requires sudo)
# For x86_64/amd64 systems:
wget -O - https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-linux-amd64-install.sh | sh

# For ARM64 systems:
wget -O - https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-linux-arm64-install.sh | sh

# Download and install to custom directory
# For x86_64/amd64 systems:
INSTALL_DIR=/opt/pre-push wget -O - https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-linux-amd64-install.sh | sh

# For ARM64 systems:
INSTALL_DIR=/opt/pre-push wget -O - https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-linux-arm64-install.sh | sh

# Or download first, then install
# For x86_64/amd64 systems:
wget https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-linux-amd64-install.sh
chmod +x pre-push-linux-amd64-install.sh
./pre-push-linux-amd64-install.sh /usr/local/bin

# For ARM64 systems:
wget https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-linux-arm64-install.sh
chmod +x pre-push-linux-arm64-install.sh
./pre-push-linux-arm64-install.sh /usr/local/bin
```

## Method 2: Manual Installation from Archive

Download and extract the archive manually:

```bash
# Download and extract archive
wget https://github.com/AlexBurnes/pre-push/releases/download/v1.0.0/pre-push_1.0.0_linux_amd64.tar.gz
tar -xzf pre-push_1.0.0_linux_amd64.tar.gz
cd pre-push_1.0.0_linux_amd64

# Install using the included install.sh
./install.sh /usr/local/bin

# Or install to user directory
./install.sh ~/.local/bin
```

## Windows Installation

Windows users can use Scoop:
```bash
scoop install pre-push
```