# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Om < Formula
  desc ""
  homepage ""
  version "7.2.0"
  bottle :unneeded

  if OS.mac?
    url "https://github.com/pivotal-cf/om/releases/download/7.2.0/om-darwin-7.2.0.tar.gz"
    sha256 "5b7c9adc52680f2a0ea7d8735131d002a018263b55bf11190541198274ab4ee0"
  end
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/pivotal-cf/om/releases/download/7.2.0/om-linux-7.2.0.tar.gz"
    sha256 "f731af7c0bf9e4c393bd2d99780643ea25ad15a809a060ef5b9db59ea649d505"
  end

  def install
    bin.install "om"
  end

  test do
    system "#{bin}/om --version"
  end
end
