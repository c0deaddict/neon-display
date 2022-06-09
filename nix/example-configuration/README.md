# Example NixOS configuration

## Build an image

```bash
nix build .#images.neon
```

## Update live installation (over SSH):

```bash
nixos-rebuild --flake '.#neon' switch --target-host root@neon --build-host localhost --use-remote-sudo -L
```
