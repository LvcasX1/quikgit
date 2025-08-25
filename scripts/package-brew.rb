class Quikgit < Formula
  desc "GitHub repository manager TUI"
  homepage "https://github.com/lvcasx1/quikgit"
  url "https://github.com/lvcasx1/quikgit/archive/v1.0.0.tar.gz"
  sha256 "" # This will be filled during build
  license "MIT"
  head "https://github.com/lvcasx1/quikgit.git", branch: "main"

  depends_on "go" => :build
  depends_on "git"

  def install
    # Build from source
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/quikgit"
    
    # Install shell completions if available
    # generate_completions_from_executable(bin/"quikgit", "completion")
  end

  def caveats
    <<~EOS
      QuikGit requires GitHub authentication. On first run, you'll be prompted
      to authenticate via QR code or by visiting a URL.
      
      Configuration is stored in ~/.quikgit/
    EOS
  end

  test do
    system "#{bin}/quikgit", "--version"
    system "#{bin}/quikgit", "--help"
  end

  # Service definition (if needed for background processes)
  # service do
  #   run [opt_bin/"quikgit", "daemon"]
  #   keep_alive true
  #   log_path var/"log/quikgit.log"
  #   error_log_path var/"log/quikgit.log"
  # end
end