{ lib, buildNpmPackage, nix-gitignore, moreutils, jq }:

buildNpmPackage {
  pname = "neon-display-frontend";
  version = "0.0.5";
  src = nix-gitignore.gitignoreSource [ ] ../../frontend;

  npmDepsHash = "sha256-xxJrOmj0JZ2sZV8hqaBX/ZESicmGjkhN7Dd6Ugx3krA=";

  buildCommands = ["npm run build"];
  installPhase = ''
    cp -r dist $out
  '';
}
