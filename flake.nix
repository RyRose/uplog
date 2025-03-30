{
  description = "Flake for setting up development of uplog";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    templ.url = "github:a-h/templ";
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
              tailwindcss
              sqlc
              nodejs
              go-tools
            ];
          };
        };
    };
}
