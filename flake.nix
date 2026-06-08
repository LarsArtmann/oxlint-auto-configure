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
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      systems,
      treefmt-nix,
    }:
    let
      version = self.rev or self.dirtyRev or "dev";

      src = nixpkgs.lib.fileset.toSource {
        root = ./.;
        fileset = nixpkgs.lib.fileset.unions [
          ./go.mod
          ./go.sum
          ./cmd
          ./internal
          ./pkg
        ];
      };
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
          ...
        }:
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
            inherit version src;
            vendorHash = null;
            ldflags = [
              "-s"
              "-w"
              "-X github.com/larsartmann/oxlint-auto-configure/internal/cli.version=${version}"
            ];
            nativeCheckInputs = [ pkgs.oxlint ];
            meta = with pkgs.lib; {
              description = "Auto-generate optimal .oxlintrc.json configurations";
              homepage = "https://github.com/larsartmann/oxlint-auto-configure";
              license = licenses.mit;
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
                  makeWrapper ${pkgs.lib.getExe config.packages.default} $out/bin/oxlint-auto-configure \
                    --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.oxlint ]}
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

              GOPRIVATE = "github.com/LarsArtmann/*";
              GOWORK = "off";
            };

            ci = pkgs.mkShellNoCC {
              packages = [
                pkgs.go_1_26
                pkgs.golangci-lint
              ];

              GOWORK = "off";
            GOPRIVATE = "github.com/LarsArtmann/*";
            };          };

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
