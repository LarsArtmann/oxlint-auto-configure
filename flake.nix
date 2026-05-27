{
  description = "Auto-generate optimal .oxlintrc.json configurations for maximum type safety";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs =
    {
      self,
      nixpkgs,
      systems,
    }:
    let
      version = self.rev or self.dirtyRev or "dev";
      supportedSystems = import systems;
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      src = nixpkgs.lib.fileset.toSource {
        root = ./.;
        fileset = nixpkgs.lib.fileset.unions [
          ./go.mod
          ./go.sum
          ./cmd
          ./internal
          ./pkg
          ./vendor
        ];
      };
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
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
        }
      );

      apps = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          pkg = self.packages.${system}.default;
          wrapped =
            pkgs.runCommandLocal "oxlint-auto-configure"
              {
                nativeBuildInputs = [ pkgs.makeWrapper ];
                meta.mainProgram = "oxlint-auto-configure";
              }
              ''
                mkdir -p $out/bin
                makeWrapper ${pkgs.lib.getExe pkg} $out/bin/oxlint-auto-configure \
                  --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.oxlint ]}
              '';
        in
        {
          default = {
            type = "app";
            program = "${wrapped}/bin/oxlint-auto-configure";
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
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
        }
      );

      overlays.default = final: prev: {
        oxlint-auto-configure = self.packages.${final.stdenv.hostPlatform.system}.default;
      };

      checks = forAllSystems (
        system:
        let
          pkg = self.packages.${system}.default;
        in
        {
          build = pkg;
          test = pkg.overrideAttrs (_: {
            doCheck = true;
          });
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt);
    };
}
