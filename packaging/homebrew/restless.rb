class Restless < Formula
  desc "Terminal-first API discovery and exploration tool"
  homepage "https://github.com/bspippi1337/restless"
  url "https://github.com/bspippi1337/restless/archive/refs/tags/v0.0.0.tar.gz"
  sha256 "REPLACE_ME"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags", "-s -w", "-o", "restless", "./cmd/restless"
    bin.install "restless"
  end

  test do
    system "#{bin}/restless", "--help"
  end
end
