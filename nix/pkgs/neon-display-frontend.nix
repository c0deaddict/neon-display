{ lib, npmlock2nix, nodejs, esbuild, moreutils, jq }:

npmlock2nix.v2.build {
  src = lib.cleanSource ../../frontend;
  inherit nodejs;

  buildCommands = [ "npm run build" ];
  installPhase = ''
    cp -r dist $out
  '';

  node_modules_attrs = {
    sourceOverrides.esbuild = sourceInfo: drv: drv.overrideAttrs(old: {
      nativeBuildInputs = [ jq moreutils ];
      postPatch = ''
        jq "del(.scripts.postinstall)" package.json \
           | jq "del(.optionalDependencies)" \
           | sponge package.json
        rm bin/esbuild
      '';
      postInstall = ''
        ln -sf ${esbuild}/bin/esbuild bin/esbuild
      '';
    });
  };
}
