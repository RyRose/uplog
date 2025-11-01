{
  description = "Flake for setting up development of uplog";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    templ.url = "github:a-h/templ?ref=v0.3.960";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-parts,
      templ,
    }@inputs:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      perSystem =
        { pkgs, system, ... }:
        {
          devShells.default = pkgs.mkShell {
            buildInputs = [
              templ.packages.${system}.templ
            ];
            packages = with pkgs; [
              go
              go-swag
              go-tools
              goose
              nodejs
              sqlc
              stylua
              tailwindcss
              luajitPackages.busted
              golangci-lint
            ];
          };
        };
    };
}
