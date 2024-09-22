{ lib, buildNpmPackage, nix-gitignore, moreutils, jq }:

buildNpmPackage {
  pname = "neon-display-frontend";
  version = "0.0.4";
  src = nix-gitignore.gitignoreSource [ ] ../../frontend;

  npmDepsHash = "sha256-ce8vuwZ41csqVR1C1IW5Pwg69/nNC64qUqnZRuvvwfE=";

  buildCommands = ["npm run build"];
  installPhase = ''
    cp -r dist $out
  '';
}
