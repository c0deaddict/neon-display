{ lib, stdenv, fetchFromGitHub, cmake }:

stdenv.mkDerivation rec {
  name = "rpi_ws281x";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "jgarff";
    repo = name;
    rev = "v${version}";
    sha256 = "sha256-Push5DMNNoTUHudMZxI7OYhaKSqbckWj5v1Jnf0ltms=";
  };

  cmakeFlags = [ "-DBUILD_SHARED=off" "-DBUILD_TEST=off" "-DCMAKE_POLICY_VERSION_MINIMUM=3.5" ];

  nativeBuildInputs = [ cmake ];

  meta = with lib; {
    homepage = "https://github.com/jgarff/rpi_ws281x";
    description = "Userspace Raspberry Pi PWM library for WS281X LEDs";
    maintainers = with maintainers; [ c0deaddict ];
    license = licenses.bsd2;
  };

}
