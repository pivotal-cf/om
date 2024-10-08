# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Om < Formula
  desc ""
  homepage ""
  version "7.14.0"

  on_macos do
    on_intel do
      url "https://github.com/pivotal-cf/om/releases/download/7.14.0/om-darwin-amd64-7.14.0.tar.gz"
      sha256 "bfb90cf72e67f4be7248744d4c145ca46f2c97ad2c877afd4e3cde1d700f7be2"

      def install
        bin.install "om"
      end
    end
    on_arm do
      url "https://github.com/pivotal-cf/om/releases/download/7.14.0/om-darwin-arm64-7.14.0.tar.gz"
      sha256 "1cbf8956a0549c390de093d6d7843b003bf7d6f9eca0fabbbc152ba693c8c7ad"

      def install
        bin.install "om"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/pivotal-cf/om/releases/download/7.14.0/om-linux-amd64-7.14.0.tar.gz"
        sha256 "6869a731d82df0ad9aa3038d6d955921decd4231293ceca32a87652123918ae2"

        def install
          bin.install "om"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/pivotal-cf/om/releases/download/7.14.0/om-linux-arm64-7.14.0.tar.gz"
        sha256 "c5ea37c592785e94fd254b5f1f9c0e50b9b4f7d1a8b8b0822db55444a9d01d82"

        def install
          bin.install "om"
        end
      end
    end
  end

  test do
    system "#{bin}/om --version"
  end
end
