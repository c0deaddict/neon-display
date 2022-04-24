{ lib, buildGoModule, npmlock2nix, rpi_ws281x, nodejs-14_x }:

buildGoModule rec {
  name = "neon-display";
  version = "0.0.1";

  src = ../..;

  vendorSha256 = "sha256-C3GrJ0YLKLK7c5Gb772KhIJk823wfW68jw3EDXewRMU=";

  propagatedBuildInputs = [ rpi_ws281x ];

  subPackages = [ "cmd/hal" "cmd/display" ];

  NIX_CFLAGS_COMPILE = "-I${rpi_ws281x}/include/ws2811";
  NIX_LDFLAGS_COMPILE = "-L${rpi_ws281x}/lib";

  preBuild = let
    frontend = npmlock2nix.build {
      src = src + "/frontend";
      nodejs = nodejs-14_x;

      buildCommands = [ "npm run build" ];
      installPhase = ''
        cp -r dist $out
      '';
    };
  in ''
    cp -r ${toString frontend} frontend/dist
  '';

  meta = with lib; {
    description = "neon-display";
    homepage = "https://github.com/c0deaddict/neon-display";
    license = licenses.mit;
    maintainers = with maintainers; [ c0deaddict ];
  };
}
