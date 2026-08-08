{
  description = "Yunxiao Woodpecker CI Forge Addon";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
        version = builtins.replaceStrings ["\n"] [""] (builtins.readFile ./version);
      in rec {
        packages = rec {
          default = yunxiao-woodpecker-addon;
          yunxiao-woodpecker-addon = pkgs.buildGoModule {
            pname = "yunxiao-woodpecker-addon";
            inherit version;
            src = ./.;

            env.CGO_ENABLED = "0";

            ldflags = [
              "-s"
              "-w"
              "-X yunxiao-woodpecker-addon/pkg/version.Version=${version}"
              "-X yunxiao-woodpecker-addon/pkg/version.BuildTime=1970-01-01T00:00:00Z"
            ];

            vendorHash = "sha256-EcyNmeTstspv9WD45Sm1YO785y1/hkdRLxPxCHsi26s=";
          };
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [packages.default];
          packages = with pkgs; [
            go
            golangci-lint
            gopls
          ];
          shellHook = ''
            echo "yunxiao-woodpecker-addon dev shell (Go ${pkgs.go.version})"
          '';
        };

        checks = {
          inherit (packages) yunxiao-woodpecker-addon;
          vet =
            pkgs.runCommand "go-vet"
            {
              buildInputs = [pkgs.go];
            }
            ''
              cd ${self}
              go vet ./...
              touch $out
            '';
          test =
            pkgs.runCommand "go-test"
            {
              buildInputs = [pkgs.go];
            }
            ''
              cd ${self}
              go test ./...
              touch $out
            '';
        };

        formatter = pkgs.alejandra;

        apps.default = {
          type = "app";
          program = "${packages.default}/bin/yunxiao-woodpecker-addon";
        };
      }
    );
}
