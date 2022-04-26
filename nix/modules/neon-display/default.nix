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
      default = "${pkgs.firefox}/bin/firefox -kiosk";
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
    };

    services.cage = {
      enable = true;
      inherit (cfg) user;
      program =
        "${cfg.browser} http://localhost:${toString cfg.settings.web_port}";
    };

    systemd.services."cage-tty1".after = [ "neon-display.service" ];

    systemd.services.neon-display = {
      wantedBy = [ "multi-user.target" ];
      after = [ "neon-display-hal.service" ];
      description = "neon-display";

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.package}/bin/display -config ${configFile}";

        User = cfg.user;
        Group = cfg.group;

        # TODO; hardening
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
        # RestrictAddressFamilies = [ "AF_UNIX" ];
        # RestrictNamespaces = true;
        # RestrictRealtime = true;
        # RestrictSUIDSGID = true;
        # SystemCallFilter = [ "@system-service" "~@privileged" "~@resources" ];
        UMask = "0007"; # required to have rwx for users group on hal.sock
      };
    };
  };
}
