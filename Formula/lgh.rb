# Copyright (c) 2025 JoeGlenn1213
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

# LGH Homebrew Formula
# To use: brew tap JoeGlenn1213/tap && brew install lgh
# Or: brew install --build-from-source lgh.rb

class Lgh < Formula
  desc "Lightweight local Git hosting service with authentication - LocalGitHub"
  homepage "https://github.com/JoeGlenn1213/lgh"
  license "MIT"
  version "1.0.0"

  # For source builds
  url "https://github.com/JoeGlenn1213/lgh/archive/refs/tags/v1.0.0.tar.gz"
  # sha256 "REPLACE_WITH_ACTUAL_SHA256"

  # Binary releases (uncomment and update when releases are available)
  # on_macos do
  #   if Hardware::CPU.arm?
  #     url "https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.0/lgh-darwin-arm64"
  #     sha256 "REPLACE_WITH_ACTUAL_SHA256"
  #   else
  #     url "https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.0/lgh-darwin-amd64"
  #     sha256 "REPLACE_WITH_ACTUAL_SHA256"
  #   end
  # end

  depends_on "go" => :build
  depends_on "git"

  def install
    # Build with version info
    ldflags = %W[
      -s -w
      -X main.Version=#{version}
      -X main.BuildDate=#{Date.today}
      -X main.GitCommit=#{Utils.git_head}
    ]

    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/lgh"
  end

  def post_install
    ohai "LGH installed successfully!"
    ohai "Quick start:"
    puts "  lgh init          # Initialize LGH environment"
    puts "  lgh serve         # Start the HTTP server"
    puts "  lgh add .         # Add current directory"
    puts "  git push lgh main # Push to local GitHub!"
    puts ""
    ohai "For network sharing with authentication:"
    puts "  lgh auth setup    # Set up username/password"
    puts "  lgh serve --bind 0.0.0.0 --read-only"
  end

  def caveats
    <<~EOS
      LGH (LocalGitHub) - Lightweight local Git hosting service

      To get started:
        1. Run 'lgh init' to set up the environment
        2. Run 'lgh serve' to start the HTTP server
        3. In your project, run 'lgh add .' to register it
        4. Push with 'git push lgh main'

      For network sharing (with authentication):
        1. Run 'lgh auth setup' to configure username/password
        2. Run 'lgh serve --bind 0.0.0.0' to expose on network
        3. Clients use: git clone http://user:pass@host:9418/repo.git

      Data is stored in ~/.localgithub/
      Server runs on http://127.0.0.1:9418 by default

      Security tip: Always enable authentication when exposing to network!
    EOS
  end

  test do
    # Test version output
    assert_match "LGH (LocalGitHub) v#{version}", shell_output("#{bin}/lgh --version")
    
    # Test help command
    assert_match "LocalGitHub", shell_output("#{bin}/lgh --help")
    
    # Test auth command exists
    assert_match "auth", shell_output("#{bin}/lgh --help")
  end
end
