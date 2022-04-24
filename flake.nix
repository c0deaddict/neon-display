{
  description = "Digital signage for Raspberry Pi";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    npmlock2nix = {
      url = "github:nix-community/npmlock2nix";
      flake = false;
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs@{ self, nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
    in {
      overlay = final: prev:
        import ./nix/pkgs/default.nix {
          pkgs = final;
          npmlock2nix = final.callPackage inputs.npmlock2nix { };
        };

      nixosModules.neon-display = import ./nix/modules/neon-display;
      nixosModule = self.nixosModules.neon-display;
      packages = forAllSystems (system:
        import ./nix/pkgs/default.nix rec {
          pkgs = import nixpkgs { inherit system; };
          npmlock2nix = pkgs.callPackage inputs.npmlock2nix { };
        });
      defaultPackage =
        forAllSystems (system: self.packages.${system}.neon-display);
    };
}
