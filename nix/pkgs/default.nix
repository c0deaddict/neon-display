{ pkgs, npmlock2nix }: rec {
  rpi_ws281x = pkgs.callPackage ./rpi_ws281x.nix { };
  neon-display = pkgs.callPackage ./neon-display.nix {
    inherit npmlock2nix rpi_ws281x;
  };
}
