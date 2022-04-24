{ npmlock2nix, nodejs-14_x }:

npmlock2nix.build {
  src = ../../frontend;
  nodejs = nodejs-14_x;

  buildCommands = [ "npm run build" ];
  installPhase = ''
    cp -r dist $out
  '';
}
