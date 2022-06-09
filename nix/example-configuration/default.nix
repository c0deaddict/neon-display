{ config, pkgs, lib, modulesPath, ... }:

with lib;

{
  imports = [ ./device-tree.nix ];

  boot.loader.grub.enable = false;
  boot.kernelPackages = pkgs.linuxPackages_rpi3;

  boot.initrd.includeDefaultModules = false;
  boot.initrd.kernelModules = [ "vc4" ];
  boot.initrd.availableKernelModules =
    lib.mkForce [ "usbhid" "usb_storage" "vc4" "bcm2835_dma" "i2c_bcm2835" ];
  boot.blacklistedKernelModules =
    [ "snd_bcm2835" ]; # blacklisted for rpi-ws281x

  boot.loader.raspberryPi.enable = true;
  boot.loader.raspberryPi.version = 3;
  boot.loader.raspberryPi.uboot.enable = true;
  boot.loader.raspberryPi.uboot.configurationLimit = 5;

  # Required to interface with LED's.
  boot.kernelModules = [ "vcio" ];
  boot.kernelParams = [ "iomem=relaxed" ];

  # Include WiFi firmware (Linux 5.15+)
  hardware.enableRedistributableFirmware = true;

  powerManagement.cpuFreqGovernor = "ondemand";

  # Not sure if this is still needed.
  hardware.opengl = {
    enable = true;
    setLdLibraryPath = true;
    driSupport = true;
  };

  fileSystems = {
    "/" = {
      device = "/dev/disk/by-label/NIXOS_SD";
      fsType = "ext4";
      options = [ "noatime" ];
    };
    "/boot/firmware" = {
      device = "/dev/disk/by-label/FIRMWARE";
      fsType = "vfat";
    };
  };

  zramSwap = {
    enable = true;
    algorithm = "zstd";
  };

  boot.tmpOnTmpfs = true;

  # Preserve space.
  documentation.nixos.enable = false;
  programs.command-not-found.enable = false;
  powerManagement.enable = false;
  nix.gc.automatic = true;
  nix.gc.options = "--delete-older-than 30d";

  networking.hostName = "neon";

  environment.systemPackages = with pkgs; [
    vim
    libraspberrypi
    util-linux # for wdctl (watchdog tool)
  ];

  services.openssh = {
    enable = true;
    passwordAuthentication = true;
    permitRootLogin = "yes";
  };

  systemd.watchdog = {
    device = "/dev/watchdog";
    # Reboots if no ping has been received for 15s. NOTE that BCM2835 only
    # supports timeouts of up to 15s.
    runtimeTime = "15s";
    # Force reboot if shutdown hangs after 10m.
    rebootTime = "10m";
  };

  # Optional: WiFi secret from sops. Value in sops should be:
  #   wireless: |
  #     PSK_SSID="hunter2"
  #
  # sops.defaultSopsFile = ./secrets.yaml;
  # sops.secrets.wireless = { };
  networking.wireless = {
    enable = lib.mkForce true;
    # environmentFile = config.sops.secrets.wireless.path;
    # networks."SSID".psk = "@PSK_SSID@";
  };

  # Optional: nats password from sops.
  # sops.secrets.nats-password.owner = "display";

  users = {
    mutableUsers = false;
    users.root.password = "nixos"; # TODO: change this.
    users.display = {
      uid = 1000;
      isNormalUser = true;
      password = "raspberry";
      extraGroups = [ "video" ];
    };
  };

  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 8080 ];
  };

  # Disable WiFi powersave.
  networking.localCommands = ''
    ${pkgs.iw}/bin/iw wlan0 set power_save off
  '';

  programs.chromium = {
    enable = true;
    extraOpts = {
      BlockThirdPartyCookies = false; # Grafana needs this.
    };
  };

  services.neon-display = {
    enable = true;
    user = "display";
    group = "users";

    settings = {
      web_bind = "0.0.0.0";
      # Enable optional nats support.
      # nats = {
      #   server_url = "nats://nats:4222";
      #   username = "neon-display";
      #   password_file = config.sops.secrets.nats-password.path;
      # };
      photos_path = "/var/lib/neon-display/photos";
      # videos_path = "/var/lib/neon-display/videos";
      off_timeout = 120; # seconds
      sites = [{
        title = "Wikipedia";
        order = 1;
        url = "https://wikipedia.org";
      }];
    };
  };

  system.stateVersion = "22.05";
}
