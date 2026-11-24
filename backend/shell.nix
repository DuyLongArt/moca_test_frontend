{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "go-cloudflare-env";

  buildInputs = with pkgs; [
    # Go Toolchain & Language Server
    go
    gopls
    golangci-lint

    # Wasm Compilers & Optimizers
    tinygo       # Produces lightweight Wasm binaries (~100-300KB)
    binaryen     # Provides wasm-opt for stripping/shrinking Wasm files

    # Cloudflare Workers & Node.js Tooling
    nodejs_22
    wrangler     # Cloudflare CLI (alternatively available via npx)
    sqlite
  ];

  # Pre-configure environment flags for Go Wasm development
  shellHook = ''
    export GOPATH="$PWD/.go"
    export PATH="$GOPATH/bin:$PATH"

    echo "⚡ Cloudflare + Go Nix Shell Loaded"
    echo "  • Go:       $(go version)"
    echo "  • TinyGo:   $(tinygo version | awk '{print $3}')"
    echo "  • Wrangler: $(wrangler --version)"
    echo "  • Node:     $(node --version)"
  '';
}