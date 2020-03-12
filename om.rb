# This file was generated by GoReleaser. DO NOT EDIT.
class Om < Formula
  desc ""
  homepage ""
  version "3.2.3"
  bottle :unneeded

  if OS.mac?
    url "https://github.com/pivotal-cf/om/releases/download/3.2.3/om-darwin-3.2.3.tar.gz"
    sha256 "4699ef8d6472cbb99d01b2282c6ca814d348d18e39603ecff7ade1b526c10b94"
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/pivotal-cf/om/releases/download/3.2.3/om-linux-3.2.3.tar.gz"
      sha256 "9b27131a7839f850d224345f88dc7ef358aaca44a61c96dbbe5d54370bcdc935"
    end
  end

  def install
    bin.install "om"
  end

  test do
    system "#{bin}/om --version"
  end
end
