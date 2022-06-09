{ lib, config, pkgs, modulesPath, ... }: {
  imports = [
    "${modulesPath}/installer/sd-card/sd-image-aarch64.nix"
  ];

  boot.loader.raspberryPi.enable = lib.mkForce false;
  boot.loader.raspberryPi.uboot.enable = lib.mkForce false;

  # Compressing takes a long time on emulated aarch64.
  sdImage.compressImage = false;

  boot.supportedFilesystems = lib.mkForce [ "ext4" "vfat" ];
}
