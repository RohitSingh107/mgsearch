{
  description = "MGSearch development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = false;
          };
        };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            golangci-lint
            air
            sqlc
            just
            watchexec
            mongosh
            redis
            nodejs_20
            yarn
          ];

          shellHook = ''
            export MONGODB_PORT=''${MONGODB_PORT:-27017}
            export REDIS_PORT=''${REDIS_PORT:-6381}
            export DATABASE_URL="mongodb://localhost:$MONGODB_PORT/mgsearch"
            export REDIS_URL="redis://127.0.0.1:$REDIS_PORT/0"

            if ! command -v mongod >/dev/null 2>&1; then
              echo "⚠️  MongoDB (mongod) not found. Install MongoDB separately:"
              echo "   - Arch/CachyOS: sudo pacman -S mongodb"
              echo "   - Ubuntu/Debian: sudo apt install mongodb"
              echo "   - macOS: brew install mongodb-community"
              echo "   - Or download from: https://www.mongodb.com/try/download/community"
            fi

            echo "👉 Dev shell ready. Run 'just dev-up' to start MongoDB/Redis."
            echo "⚠️  Configure MEILISEARCH_URL / MEILISEARCH_API_KEY for your cloud host in .env"
          '';
        };

        formatter = pkgs.nixpkgs-fmt;
      });
}

