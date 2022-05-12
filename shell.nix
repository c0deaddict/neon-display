{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    nodejs-14_x esbuild
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
    exiftool
  ];
}
