{
  description = "Auto-generate optimal .oxlintrc.json configurations for maximum type safety";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "git+ssh://git@github.com/LarsArtmann/go-nix-helpers?ref=master";
      flake = false;
    };

    go-finding = {
      url = "git+ssh://git@github.com/LarsArtmann/go-finding?ref=master";
      flake = false;
    };

    gogenfilter = {
      url = "git+ssh://git@github.com/LarsArtmann/gogenfilter?ref=master";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      systems,
      treefmt-nix,
      ...
    }:
    let
      version = self.rev or self.dirtyRev or "dev";
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        let
          mkPreparedSource = import (inputs.go-nix-helpers + "/mkPreparedSource.nix") {
            inherit pkgs lib;
            goPkg = pkgs.go_1_26;
          };

          preparedSrc = mkPreparedSource {
            name = "oxlint-auto-configure";
            inherit version;
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.unions [
                ./go.mod
                ./go.sum
                ./cmd
                ./internal
                ./pkg
              ];
            };
            deps = {
              "github.com/larsartmann/go-finding" = inputs.go-finding;
              "github.com/LarsArtmann/gogenfilter/v3" = inputs.gogenfilter;
            };
          };

          src = preparedSrc;

          # To update after a dependency change: `nix build .#default`, then
          # copy the `got:` sha256 from the hash-mismatch error below.
          vendorHash = "sha256-Df8bY/y+NCQgX57xAr9wyIP4kAXfd4xZyPksjk1xpHg=";
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              nixfmt.enable = true;
            };
          };

          packages.default = pkgs.buildGoModule {
            pname = "oxlint-auto-configure";
            inherit version src vendorHash;
            proxyVendor = false;
            ldflags = [
              "-s"
              "-w"
              "-X github.com/larsartmann/oxlint-auto-configure/internal/cli.version=${version}"
            ];
            nativeCheckInputs = [ pkgs.oxlint ];
            env.GOEXPERIMENT = "jsonv2";
            meta = with lib; {
              description = "Auto-generate optimal .oxlintrc.json configurations";
              homepage = "https://github.com/larsartmann/oxlint-auto-configure";
              license = licenses.mit;
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
              mainProgram = "oxlint-auto-configure";
            };
          };

          apps.default = {
            type = "app";
            program = "${
              pkgs.runCommandLocal "oxlint-auto-configure"
                {
                  nativeBuildInputs = [ pkgs.makeWrapper ];
                  meta.mainProgram = "oxlint-auto-configure";
                }
                ''
                  mkdir -p $out/bin
                  makeWrapper ${lib.getExe config.packages.default} $out/bin/oxlint-auto-configure \
                    --prefix PATH : ${lib.makeBinPath [ pkgs.oxlint ]}
                ''
            }/bin/oxlint-auto-configure";
          };

          devShells = {
            default = pkgs.mkShell {
              packages = with pkgs; [
                go_1_26
                gopls
                gotools
                golangci-lint
                oxlint
              ];

              GOPRIVATE = "github.com/larsartmann/*,github.com/LarsArtmann/*";
              GOWORK = "off";
              GOEXPERIMENT = "jsonv2";
            };

            ci = pkgs.mkShellNoCC {
              packages = [
                pkgs.go_1_26
                pkgs.golangci-lint
              ];

              GOWORK = "off";
              GOPRIVATE = "github.com/larsartmann/*,github.com/LarsArtmann/*";
              GOEXPERIMENT = "jsonv2";
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;
            test = config.packages.default.overrideAttrs (_: {
              doCheck = true;
            });
          };
        };

      flake.overlays.default = final: _prev: {
        oxlint-auto-configure = self.packages.${final.stdenv.hostPlatform.system}.default;
      };
    };
}
