{ lib, buildGoModule, nix-gitignore, rpi_ws281x, neon-display-frontend }:

buildGoModule rec {
  name = "neon-display";
  version = "0.0.5";

  src = nix-gitignore.gitignoreSource [ ] ../..;

  vendorHash = "sha256-HCJVaEzS8yRpJ/UwJxpbNV7xqPZO0IjLwaCXVodVemg=";

  propagatedBuildInputs = [ rpi_ws281x ];

  subPackages = [ "cmd/hal" "cmd/display" ];

  NIX_CFLAGS_COMPILE = "-I${rpi_ws281x}/include/ws2811";
  NIX_LDFLAGS_COMPILE = "-L${rpi_ws281x}/lib";

  preBuild = ''
    cp -r ${neon-display-frontend} frontend/dist
  '';

  meta = with lib; {
    description = "neon-display";
    homepage = "https://github.com/c0deaddict/neon-display";
    license = licenses.mit;
    maintainers = with maintainers; [ c0deaddict ];
  };
}
