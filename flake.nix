{
  description = "Auto-generate optimal .oxlintrc.json configurations for maximum type safety";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers/29e39b25b691ce51c0daa7be85f551c7016ed2c5";
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
      url = "github:LarsArtmann/go-finding/v1.10.0";
      flake = false;
    };

    linter-autoconfigure-sdk = {
      url = "github:LarsArtmann/linter-autoconfigure-sdk/v0.2.0";
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
        vendorHash = "sha256-zgK/kRZb0CdaXpvjnZ76hxrsjTzTI9YQjSebpu5Oxhc=";
        description = "Auto-generate optimal .oxlintrc.json configurations";
        enableCheck = false;
        subPackages = [ "cmd/oxlint-auto-configure" ];

        # go.mod floor (go 1.27.1) is newer than nixpkgs' go (1.26.7); the
        # tarball toolchain keeps the sandboxed treefmt/goimports check
        # offline-safe (matching go on PATH + GOTOOLCHAIN=local).
        goTarballVersion = "1.27.1";
        goTarballHash = "sha256-TkCKuuEm2Ra2FkYnGT8sVPDjyhMS1pO4bbRfhiqyOLE=";

        deps = {
          "github.com/larsartmann/go-atomic-write" = inputs.go-atomic-write;
          "github.com/larsartmann/go-error-family" = inputs.go-error-family;
          "github.com/larsartmann/go-finding" = inputs.go-finding;
          "github.com/larsartmann/linter-autoconfigure-sdk" = inputs.linter-autoconfigure-sdk;
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
