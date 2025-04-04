# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class SdLocal < Formula
  desc "Screwdriver local mode."
  homepage "https://github.com/screwdriver-cd/sd-local"
  version "1.0.58"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/screwdriver-cd/sd-local/releases/download/v1.0.58/sd-local_darwin_amd64"
      sha256 "6b6d375bd3e8259b1311a94d43d7f8788ad34c78714fdc625f3bf3b56e960094"

      def install
        bin.install File.basename(@stable.url) => "sd-local"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/screwdriver-cd/sd-local/releases/download/v1.0.58/sd-local_darwin_arm64"
      sha256 "f6a15d2cbb000fa6dd10cca7389316531f48c460702ddebedf557fb34c91b6ac"

      def install
        bin.install File.basename(@stable.url) => "sd-local"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/screwdriver-cd/sd-local/releases/download/v1.0.58/sd-local_linux_amd64"
        sha256 "6de1f89ad4981333e2b33b42460eaa1f31ef68048b04d0b66bb92a7e8d9eb6f2"

        def install
          bin.install File.basename(@stable.url) => "sd-local"
        end
      end
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/screwdriver-cd/sd-local/releases/download/v1.0.58/sd-local_linux_arm64"
        sha256 "20b8a5c6730971925212eb8578d2dc4d314600790fddb3a7e06bb24a00caaab4"

        def install
          bin.install File.basename(@stable.url) => "sd-local"
        end
      end
    end
  end

  test do
    system "#{bin}/sd-local", "--help"
  end
end
