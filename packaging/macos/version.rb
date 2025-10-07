class PrePush < Formula
  desc "Cross-platform Git pre-push hook runner with DAG-based execution"
  homepage "https://github.com/AlexBurnes/pre-push"
  url "https://github.com/AlexBurnes/pre-push/releases/download/v1.8.2/pre-push-1.4.5-darwin-amd64.tar.gz"
  version "v1.4.7"
  sha256 "PLACEHOLDER_SHA256"
  license "Apache-2.0"
  
  if Hardware::CPU.arm?
    url "https://github.com/AlexBurnes/pre-push/releases/download/v1.8.2/pre-push-1.4.5-darwin-arm64.tar.gz"
    sha256 "PLACEHOLDER_SHA256_ARM64"
  end

  def install
    bin.install "pre-push"
    man1.install "pre-push.1" if File.exist?("pre-push.1")
  end

  test do
    assert_match "pre-push", shell_output("#{bin}/pre-push --version").strip
    assert_match "Usage:", shell_output("#{bin}/pre-push --help").strip
  end
end