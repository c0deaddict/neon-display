{ lib, buildNpmPackage, nix-gitignore, moreutils, jq }:

buildNpmPackage {
  pname = "neon-display-frontend";
  version = "0.0.5";
  src = nix-gitignore.gitignoreSource [ ] ../../frontend;

  npmDepsHash = "sha256-0PGlX/8ETxfWQcCzLwR+JWXOUm3jQ10aNu7tvOZ+tzA=";

  buildCommands = ["npm run build"];
  installPhase = ''
    cp -r dist $out
  '';
}
