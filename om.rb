# This file was generated by GoReleaser. DO NOT EDIT.
class Om < Formula
  desc ""
  homepage ""
  version "5.0.0"
  bottle :unneeded

  if OS.mac?
    url "https://github.com/pivotal-cf/om/releases/download/5.0.0/om-darwin-5.0.0.tar.gz"
    sha256 "3224ed7d93318f843248aaa921a9b5f4ebc6c596833d3eb4d4b50b69a30c29ec"
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/pivotal-cf/om/releases/download/5.0.0/om-linux-5.0.0.tar.gz"
      sha256 "9a1f099a9e8252de36a7f581e6c0139996e10dae89a9286592cd7dfcba79a8df"
    end
  end

  def install
    bin.install "om"
  end

  test do
    system "#{bin}/om --version"
  end
end
