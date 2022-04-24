{ config, pkgs, lib, ... }:

with lib;

let

  cfg = config.services.neon-display;
  configFile = {}; # TODO make json config

in
{
  options.services.neon-display = {
    enable = mkEnableOption "neon-display";

    user = mkOption {
      type = types.string;
      example = "display";
    };

    package = mkOption {
      type = types.package;
      default = pkgs.neon-display;
    };
  };

  config = mkIf cfg.enable {
    services.cage = {
      enable = true;
      inherit (cfg) user;
      program = "${pkgs.neon-display}/bin/display -config ${configFile}";
    };

    systemd.services.neon-display-hal = {
      description = "neon-display hardware abstraction layer";

      # TODO: make cage depend on this service.
      serviceConfig = {
        Type = "oneshot";
        PrivateTmp = true;
        # TODO: use /run dir created by systemd.
        # TODO: tighten security a bit more?
        ExecStart = "${pkgs.neon-display}/bin/hal -hal-socket-path /run/hal.sock";
      };
    };
  };
}
