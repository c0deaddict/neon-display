{ pkgs }: rec {
  rpi_ws281x = pkgs.callPackage ./rpi_ws281x.nix { };
  neon-display = pkgs.callPackage ./neon-display.nix { inherit rpi_ws281x neon-display-frontend; };
  neon-display-frontend = pkgs.callPackage ./neon-display-frontend.nix { };
}
