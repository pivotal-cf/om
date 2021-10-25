# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Om < Formula
  desc ""
  homepage ""
  version "7.3.2"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/pivotal-cf/om/releases/download/7.3.2/om-darwin-7.3.2.tar.gz"
      sha256 "7a78819b4698c4a5337de5d3c77730c23cebdc7c7dd502f93679914ab9c6b08a"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/pivotal-cf/om/releases/download/7.3.2/om-linux-7.3.2.tar.gz"
      sha256 "352965ec4d8be070e0e00baf4fec77eaaf20c6590acb00740c2d378ecb44dfb9"
    end
  end

  def install
    bin.install "om"
  end

  test do
    system "#{bin}/om --version"
  end
end
