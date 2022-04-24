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
      type = types.string;
      example = "display";
    };

    group = mkOption {
      type = types.string;
      example = "users";
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
      hal_socket_path = "/var/run/neon-display/hal.sock";
    };

    services.cage = {
      enable = true;
      inherit (cfg) user;
      program = "${cfg.package}/bin/display -config ${configFile}";
    };

    systemd.services."cage-tty1".after = [ "neon-display-hal.service" ];

    systemd.services.neon-display-hal = {
      description = "neon-display hardware abstraction layer";

      serviceConfig = {
        Type = "simple";
        ExecStart =
          "${cfg.package}/bin/hal -hal-socket-path ${cfg.settings.hal_socket_path}";

        User = "root";
        Group = cfg.group;

        RuntimeDirectoryMode = "0750";
        RuntimeDirectory = "neon-display";

        # Hardening
        CapabilityBoundingSet = "";
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        NoNewPrivileges = true;
        PrivateTmp = true;
        PrivateUsers = true;
        ProcSubset = "pid";
        ProtectClock = true;
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectProc = "invisible";
        ProtectSystem = "strict";
        ReadOnlyPaths = [ ];
        ReadWritePaths = [ ];
        RestrictAddressFamilies = [ "AF_UNIX" ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallFilter = [ "@system-service" "~@privileged" "~@resources" ];
        UMask = "0077";
      };
    };
  };
}
