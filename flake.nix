{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};

        app = pkgs.buildGoModule {
          pname = "short-form";
          version = "0.1.0";

          env.CGO_ENABLED = 0;

          src = ./.;

          vendorHash = "sha256-MD8RJaVGnYj6nQ6Xgq8AthgFY+r/cuCmssOyfNrKC1s=";

          nativeBuildInputs = [
            pkgs.ffmpeg
          ];

          preBuild = ''
          '';

          meta = with pkgs.lib; {
            description = "short form video discord app";
            license = licenses.bsd3;
          };
        };
      in {
        packages = {
          default = app;
          app = app;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_25
            gopls
            golangci-lint

            air
            # swag
            pre-commit
          ];

          shellHook = ''
            echo "Go development environment ready"
            echo "Go version: $(go version)"
          '';
        };
      }
    );
}
