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
            postgresql_16
            redis
            nodejs_20
            yarn
          ];

          shellHook = ''
            export PGPORT=''${PGPORT:-5544}
            export REDIS_PORT=''${REDIS_PORT:-6381}
            export DATABASE_URL="postgres://mgsearch:mgsearch@localhost:$PGPORT/mgsearch?sslmode=disable"
            export REDIS_URL="redis://127.0.0.1:$REDIS_PORT/0"

            echo "👉 Dev shell ready. Run 'just dev-up' to start Postgres/Redis."
            echo "⚠️  Configure MEILISEARCH_URL / MEILISEARCH_API_KEY for your cloud host in .env"
          '';
        };

        formatter = pkgs.nixpkgs-fmt;
      });
}

