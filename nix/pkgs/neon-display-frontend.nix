{ lib, buildNpmPackage, nix-gitignore, moreutils, jq }:

buildNpmPackage {
  pname = "neon-display-frontend";
  version = "0.0.5";
  src = nix-gitignore.gitignoreSource [ ] ../../frontend;

  npmDepsHash = "sha256-l4h1LDIG0rBCGOgkckrKQkbzV3wA5JJp9YnSQP660XI=";

  buildCommands = ["npm run build"];
  installPhase = ''
    cp -r dist $out
  '';
}
