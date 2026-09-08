{
  description = "Auto-generate optimal .oxlintrc.json configurations for maximum type safety";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers/a97742e806193cd7e4c457439c7e117a6cfd1fe7";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-atomic-write = {
      url = "github:LarsArtmann/go-atomic-write/v0.5.1";
      flake = false;
    };

    go-error-family = {
      url = "github:LarsArtmann/go-error-family/v0.10.0";
      flake = false;
    };

    go-finding = {
      url = "github:LarsArtmann/go-finding/v1.8.0";
      flake = false;
    };

    gogenfilter = {
      url = "github:LarsArtmann/gogenfilter/v3.4.0";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      ...
    }:
    let
      version = self.rev or self.dirtyRev or "dev";
      commit = self.shortRev or self.dirtyShortRev or "unknown";
      date = builtins.substring 0 8 (self.lastModifiedDate or "19700101");
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "oxlint-auto-configure";
        vendorHash = "sha256-wlHMTO7iYF3/esOhafI2B9lCo+sYL8erRbIEjFnEwAQ=";
        description = "Auto-generate optimal .oxlintrc.json configurations";
        enableCheck = false;

        deps = {
          "github.com/larsartmann/go-atomic-write" = inputs.go-atomic-write;
          "github.com/larsartmann/go-error-family" = inputs.go-error-family;
          "github.com/larsartmann/go-finding" = inputs.go-finding;
          "github.com/LarsArtmann/gogenfilter/v3" = inputs.gogenfilter;
        };

        src = inputs.nixpkgs.lib.fileset.toSource {
          root = ./.;
          fileset = inputs.nixpkgs.lib.fileset.unions [
            ./go.mod
            ./go.sum
            ./cmd
            ./internal
            ./pkg
          ];
        };

        ldflags = [
          "-s"
          "-w"
          "-X github.com/larsartmann/oxlint-auto-configure/internal/cli.version=${version}"
          "-X github.com/larsartmann/oxlint-auto-configure/internal/cli.commit=${commit}"
          "-X github.com/larsartmann/oxlint-auto-configure/internal/cli.date=${date}"
          "-X github.com/larsartmann/oxlint-auto-configure/internal/cli.builtBy=nix"
        ];

        extraBuildAttrs.preBuild = "export GOEXPERIMENT=jsonv2";

        shellExtraEnv = {
          GOEXPERIMENT = "jsonv2";
        };

        devShellExtraPackages = pkgs: [
          pkgs.oxlint
          pkgs.gopls
          pkgs.gotools
          pkgs.golangci-lint
        ];

        enableNixfmt = true;
      };

      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        {
          checks.test = config.packages.default.overrideAttrs (_old: {
            doCheck = true;
            nativeCheckInputs = [ pkgs.oxlint ];
          });

          apps.default = lib.mkForce {
            type = "app";
            program = "${
              pkgs.runCommandLocal "oxlint-auto-configure-wrapped"
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
        };
    };
}
