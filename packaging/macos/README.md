# macOS Homebrew Tap for pre-push CLI

This directory contains the Homebrew formula for the pre-push CLI utility.

## Installation

To install the pre-push CLI via Homebrew:

```bash
# Add the tap
brew tap AlexBurnes/homebrew-tap https://github.com/AlexBurnes/homebrew-tap

# Install the formula
brew install pre-push
```

## Updating

To update to the latest version:

```bash
brew update
brew upgrade pre-push
```

## Uninstalling

To remove the pre-push CLI:

```bash
brew uninstall pre-push
```

## Formula Details

- **Formula Name**: `pre-push`
- **Description**: Cross-platform Git pre-push hook runner with DAG-based execution
- **License**: Apache-2.0
- **Homepage**: https://github.com/AlexBurnes/pre-push
- **Architectures**: amd64, arm64 (Apple Silicon)

## Manual Installation

If you prefer not to use Homebrew, you can download the binary directly:

```bash
# For Intel Macs
curl -L https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-darwin-amd64.tar.gz | tar -xz
sudo mv pre-push /usr/local/bin/

# For Apple Silicon Macs
curl -L https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-darwin-arm64.tar.gz | tar -xz
sudo mv pre-push /usr/local/bin/
```

## Using Installer Scripts

You can also use the self-extracting installer scripts:

```bash
# For Intel Macs
wget -O - https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-darwin-amd64-install.sh | sh

# For Apple Silicon Macs
wget -O - https://github.com/AlexBurnes/pre-push/releases/latest/download/pre-push-darwin-arm64-install.sh | sh
```

## Verification

After installation, verify the installation:

```bash
pre-push --version
pre-push --help
```

## Support

For issues and support, please visit the [main project repository](https://github.com/AlexBurnes/pre-push).