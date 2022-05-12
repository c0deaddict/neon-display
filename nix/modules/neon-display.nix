{ config, pkgs, lib, ... }:

with lib;

let

  cfg = config.services.neon-display;

  format = pkgs.formats.json { };
  configFile = format.generate "config.json" cfg.settings;

in {
  options.services.neon-display = {
    enable = mkEnableOption "neon-display";

    user = mkOption {
      type = types.str;
      example = "display";
    };

    group = mkOption {
      type = types.str;
      example = "users";
    };

    browser = mkOption {
      type = types.str;
      # Firefox hardware acceleration is not working properly on Raspberry
      # Pi's. Without it Grafana is way too heavy, pushing the CPU temperature
      # over the limit of 85C.
      #
      # Chromium is run in XWayland. Ozone with --use-gl does not work
      # currently.  These flags are used to enable HW acceleration on Raspberry
      # Pi's.
      #
      # Video acceleration is not working yet, Chromium needs to be compiled
      # with V4L2 which is not enabled by default.
      default = let
        flags = [
          "--ignore-gpu-blocklist"
          "--enable-gpu-rasterization"
          "--enable-zero-copy"
          "--enable-drdc"
          "--canvas-oop-rasterization"
          "--enable-accelerated-video-decode"
          "--use-gl=egl"
          "--enable-features=VaapiVideoDecoder"
          "--disable-features=UseChromeOSDirectVideoDecoder"
          "--remote-debugging-port=9222"
          "--kiosk"
        ];
      in "${pkgs.ungoogled-chromium}/bin/chromium ${
        concatStringsSep " " flags
      }";
      example = literalExpression ''"''${pkgs.firefox}/bin/firefox -kiosk"'';
    };

    package = mkOption {
      type = types.package;
      default = pkgs.neon-display;
    };

    settings = mkOption {
      default = { };
      type = format.type;
    };
  };

  config = mkIf cfg.enable {
    services.neon-display.settings = {
      web_port = 8080;
      hal_socket_path = "/var/run/neon-display/hal.sock";
      cache_path = "/var/lib/neon-display/cache";
    };

    services.cage = {
      enable = true;
      inherit (cfg) user;
      program =
        "${cfg.browser} http://localhost:${toString cfg.settings.web_port}";
    };

    # Hide the cursor if no input devices are connected.
    systemd.services."cage-tty1" = {
      environment = {
        WLR_LIBINPUT_NO_DEVICES = "1";
        NO_AT_BRIDGE = "1";
      };
    };

    systemd.services."cage-tty1".after = [ "neon-display.service" ];

    systemd.services.neon-display = {
      wantedBy = [ "multi-user.target" ];
      wants = [ "network.target" ];
      after = [ "network.target" "neon-display-hal.service" ];
      description = "neon-display";

      path = [ pkgs.exiftool ];

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.package}/bin/display -config ${configFile}";
        Restart = "on-failure";

        User = cfg.user;
        Group = cfg.group;

        StateDirectory = "neon-display";

        # Hardening
        TemporaryFileSystem = "/:ro";
        BindReadOnlyPaths = [
          "/nix/store"
          "-/etc/resolv.conf"
          "-/etc/nsswitch.conf"
          "-/etc/hosts"
          "-/etc/localtime"
          cfg.settings.hal_socket_path
        ] ++ (lib.optional (cfg.settings ? photos_path)
          cfg.settings.photos_path)
          ++ (lib.optional (cfg.settings ? videos_path)
            cfg.settings.videos_path);

        CapabilityBoundingSet = "";
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        NoNewPrivileges = true;
        PrivateDevices = true;
        PrivateTmp = true;
        PrivateUsers = true;
        ProcSubset = "pid";
        ProtectClock = true;
        ProtectControlGroups = true;
        # Does not play well with TemporaryFileSystem.
        # ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        # Using temporary filesystem instead of this.
        # ProtectSystem = "strict";
        ProtectProc = "invisible";
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallFilter = [ "@system-service" "~@privileged" "~@resources" ];
        UMask = "0077";
      };
    };

    systemd.services.neon-display-hal = {
      wantedBy = [ "multi-user.target" ];
      description = "neon-display hardware abstraction layer";

      path = [ pkgs.libraspberrypi ];

      serviceConfig = {
        Type = "simple";
        ExecStart =
          "${cfg.package}/bin/hal -hal-socket-path ${cfg.settings.hal_socket_path}";
        Restart = "on-failure";

        User = "root";
        Group = cfg.group;

        RuntimeDirectoryMode = "0750";
        RuntimeDirectory = "neon-display";

        # Hardening
        # CapabilityBoundingSet = "";
        # LockPersonality = true;
        # MemoryDenyWriteExecute = true;
        # NoNewPrivileges = true;
        # PrivateTmp = true;
        # PrivateUsers = true;
        # ProcSubset = "pid";
        # ProtectClock = true;
        # ProtectHome = true;
        # ProtectHostname = true;
        # ProtectKernelLogs = true;
        # ProtectKernelModules = true;
        # ProtectKernelTunables = true;
        # ProtectProc = "invisible";
        # ProtectSystem = "strict";
        # ReadOnlyPaths = [ ];
        # ReadWritePaths = [ ];
        RestrictAddressFamilies = [ "AF_UNIX" ];
        # RestrictNamespaces = true;
        # RestrictRealtime = true;
        # RestrictSUIDSGID = true;
        # SystemCallFilter = [ "@system-service" "~@privileged" "~@resources" ];
        UMask = "0007"; # required to have rwx for users group on hal.sock
      };
    };
  };
}
