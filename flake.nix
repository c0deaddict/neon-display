{
  description = "Digital signage for Raspberry Pi";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
  };

  outputs = inputs@{ self, nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
    in
    {
      overlay = final: prev: import ./nix/pkgs/default.nix { pkgs = final; };

      nixosModules.neon-display = import ./nix/modules/neon-display.nix;
      nixosModule = self.nixosModules.neon-display;
      packages = forAllSystems (system:
        import ./nix/pkgs/default.nix rec {
          pkgs = import nixpkgs { inherit system; };
        });
      defaultPackage = forAllSystems (system: self.packages.${system}.neon-display);

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            nativeBuildInputs = [ pkgs.bashInteractive ];
            buildInputs = with pkgs; [
              go_1_21
              nodejs
              esbuild
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              exiftool
            ];
          };
        });
    };
}
