{ config, pkgs, lib, ... }:

{
  # Custom implementation for the device tree that uses Raspberry Pi's dtoverlay
  # command to merge overlays into the DTB files. This also fixes the issue of
  # U-Boot not loading the correct DTB. This issue is "fixed" in NixOS by this code:
  #   https://github.com/NixOS/nixpkgs/blob/7beebb590d541ffa534aa34b0b163f81c3c72c2c/pkgs/os-specific/linux/kernel/linux-rpi.nix#L57
  #
  # Unfortunately the fix in NixOS doesn't work if one uses
  # hardware.deviceTree.overlays. Then the fixup is ignored because DTB's are
  # compiled from source.
  #
  # Besides, the `compatible` fields in the overlays are not correctly
  # interpreted (by fdtoverlay) or are just plain wrong (in the
  # raspberrypifw). It looks like dtoverlay just ignores this field altogether.
  hardware.deviceTree = let
    compileDTS = name: f:
      pkgs.runCommand "${name}.dtbo" { } ''
        ${pkgs.dtc}/bin/dtc -I dts ${f} -O dtb -@ -o $out
      '';

    withDTBOs = xs:
      lib.flip map xs (o:
        o // {
          dtboFile = if isNull o.dtboFile then
            if !isNull o.dtsFile then
              compileDTS o.name o.dtsFile
            else
              compileDTS o.name (pkgs.writeText "dts" o.dtsText)
          else
            o.dtboFile;
        });
  in {
    enable = false;
    # TODO: add support for dtparam=key=value:
    # dtmerge base.dtb out.dtb - key=value
    package = pkgs.runCommand "device-tree" {
      nativeBuildInputs = with pkgs; [ dtc libraspberrypi ];
    } ''
      dtbDir=$out/broadcom
      mkdir -p $dtbDir
      cp -v ${pkgs.raspberrypifw}/share/raspberrypi/boot/*.dtb $dtbDir/

      for dtb in $(find $dtbDir -type f -name '*.dtb'); do
        ${
          lib.flip (lib.concatMapStringsSep "\n")
          (withDTBOs config.hardware.deviceTree.overlays) (o: ''
            echo "Applying overlay ${o.name} to $(basename $dtb)"
            mv $dtb{,.in}
            dtmerge "$dtb.in" "$dtb" ${o.dtboFile}
            rm "$dtb.in"
          '')
        }
      done

      # https://github.com/NixOS/nixpkgs/blob/7beebb590d541ffa534aa34b0b163f81c3c72c2c/pkgs/os-specific/linux/kernel/linux-rpi.nix#L57
      # Make copies of the DTBs named after the upstream names so that U-Boot finds them.
      # This is ugly as heck, but I don't know a better solution so far.
      copyDTB() {
        cp -v "$dtbDir/$1" "$dtbDir/$2"
      }
      copyDTB bcm2708-rpi-zero-w.dtb bcm2835-rpi-zero.dtb
      copyDTB bcm2708-rpi-zero-w.dtb bcm2835-rpi-zero-w.dtb
      copyDTB bcm2708-rpi-b.dtb bcm2835-rpi-a.dtb
      copyDTB bcm2708-rpi-b.dtb bcm2835-rpi-b.dtb
      copyDTB bcm2708-rpi-b.dtb bcm2835-rpi-b-rev2.dtb
      copyDTB bcm2708-rpi-b-plus.dtb bcm2835-rpi-a-plus.dtb
      copyDTB bcm2708-rpi-b-plus.dtb bcm2835-rpi-b-plus.dtb
      copyDTB bcm2708-rpi-b-plus.dtb bcm2835-rpi-zero.dtb
      copyDTB bcm2708-rpi-cm.dtb bcm2835-rpi-cm.dtb
      copyDTB bcm2709-rpi-2-b.dtb bcm2836-rpi-2-b.dtb
      copyDTB bcm2710-rpi-zero-2.dtb bcm2837-rpi-zero-2.dtb
      copyDTB bcm2710-rpi-3-b.dtb bcm2837-rpi-3-b.dtb
      copyDTB bcm2710-rpi-3-b-plus.dtb bcm2837-rpi-3-a-plus.dtb
      copyDTB bcm2710-rpi-3-b-plus.dtb bcm2837-rpi-3-b-plus.dtb
      copyDTB bcm2710-rpi-cm3.dtb bcm2837-rpi-cm3.dtb
      copyDTB bcm2711-rpi-4-b.dtb bcm2838-rpi-4-b.dtb
    '';

    overlays = let cma = 256;
    in [
      {
        # https://github.com/matthewbauer/nixiosk/blob/6ce28320dc425812c8b287a82cb100b3e90b60a8/hardware/raspberrypi.nix#L71
        name = "cma";
        dtsText = ''
          /dts-v1/;
          /plugin/;
          / {
            compatible = "brcm,bcm";
            fragment@0 {
              target = <&cma>;
              __overlay__ {
                size = <(${toString cma} * 1024 * 1024)>;
              };
            };
          };
        '';
      }
      # TODO: support: dtparam=sd_poll_once=on
      {
        name = "sd_poll_once";
        dtsText = ''
          /dts-v1/;
          /plugin/;
          / {
            compatible = "brcm,bcm";
            fragment@0 {
              target = <&sdhost>;
              __overlay__ {
                non-removable;
              };
            };
          };
        '';
      }
      # TODO support: dtparam=watchdog=on
      {
        name = "watchdog";
        dtsText = ''
          /dts-v1/;
          /plugin/;
          / {
            compatible = "brcm,bcm";
            fragment@0 {
              target = <&watchdog>;
              __overlay__ {
                status = "okay";
              };
            };
          };
        '';
      }
      {
        name = "vc4-fkms-v3d";
        dtboFile =
          "${pkgs.raspberrypifw}/share/raspberrypi/boot/overlays/vc4-fkms-v3d.dtbo";
      }
      # Normally the Pi firmware will fill this in. Because of Uboot this
      # information is lost. rpi_ws281x requires this identification to work.
      {
        name = "system";
        dtsText = ''
          /dts-v1/;
          /plugin/;

          / {
              compatible = "raspberrypi,3-model-b\0brcm,bcm2837";

              fragment@0 {
                  target-path = "/";
                  __overlay__ {
                      system {
                          linux,serial = <0x00 0xc7e96ee3>;
                          linux,revision = <0xa02082>;
                      };
                  };
              };
          };
        '';
      }
    ];
  };
}
