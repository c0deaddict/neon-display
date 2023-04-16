{ npmlock2nix, nodejs-14_x }:

npmlock2nix.v1.build {
  src = ../../frontend;
  nodejs = nodejs-14_x;

  buildCommands = [ "npm run build" ];
  installPhase = ''
    cp -r dist $out
  '';
}
