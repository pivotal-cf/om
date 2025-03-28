# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Om < Formula
  desc ""
  homepage ""
  version "7.15.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/pivotal-cf/om/releases/download/7.15.0/om-darwin-amd64-7.15.0.tar.gz"
      sha256 "5a3959ac5997e1d5b356c160a4cb9681f84f596d9117398c545dcd0718581627"

      def install
        bin.install "om"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/pivotal-cf/om/releases/download/7.15.0/om-darwin-arm64-7.15.0.tar.gz"
      sha256 "f62417bc6ff206d2b890761f6a6ec86b48edd800ee00bd8fa8607fbec003f376"

      def install
        bin.install "om"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/pivotal-cf/om/releases/download/7.15.0/om-linux-amd64-7.15.0.tar.gz"
        sha256 "ea2180e1e41fe2ec292f9ad327162818864deb711615f7b7886adafcefe799d5"

        def install
          bin.install "om"
        end
      end
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/pivotal-cf/om/releases/download/7.15.0/om-linux-arm64-7.15.0.tar.gz"
        sha256 "a7272dbf5bcc1dab5a59bac1e90d433ba039a3ba30b17a33b5d16ecafaf9f11d"

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
