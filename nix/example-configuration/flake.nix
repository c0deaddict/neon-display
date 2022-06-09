{
  description = "NixOS configuration with flakes";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    sops-nix = {
      url = "github:Mic92/sops-nix/master";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    neon-display = {
      url = "github:c0deaddict/neon-display";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs@{ self, nixpkgs, ... }:
    let
      nixosSystem = args:
        let pkgs = import nixpkgs { inherit (args) system; };
        in import (pkgs.path + /nixos/lib/eval-config.nix) (args // {
          modules = args.modules ++ [
            { config._module.args = { inherit (self) modulesPath; }; }
            {
              system.nixos.versionSuffix = ".${
                  pkgs.lib.substring 0 8
                  (self.lastModifiedDate or self.lastModified or "19700101")
                }.${self.shortRev or "dirty"}";
              system.nixos.revision = pkgs.lib.mkIf (self ? rev) self.rev;
            }
          ];
        });
    in rec {
      nixosConfigurations = {
        neon = nixosSystem {
          system = "aarch64-linux";
          modules = [
            ./.
            inputs.sops-nix.nixosModules.sops
            inputs.neon-display.nixosModule
            ({ pkgs, ... }:
              let
                crossPkgs = import nixpkgs {
                  localSystem.system = "x86_64-linux";
                  crossSystem.system = "aarch64-linux";
                  overlays = [
                    inputs.sops-nix.overlay
                    # Lazy cross compiling.
                    (final: prev: {
                      # for sops-install-secrets
                      inherit (pkgs) go;
                    })
                  ];
                };
              in {
                nixpkgs.overlays = [ inputs.neon-display.overlay ];
                sops.package = crossPkgs.sops-install-secrets;
                services.neon-display.package = pkgs.neon-display;
              })
            {
              nix.nixPath = [ "nixpkgs=${nixpkgs}" ];
              nix.registry = { nixpkgs = { flake = nixpkgs; }; };
            }
          ];
        };
      };

      images = {
        neon = (nixosConfigurations.neon.extendModules {
          modules = [ ./sd-image.nix ];
        }).config.system.build.sdImage;
      };
    };
}
