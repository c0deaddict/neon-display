{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    nodejs esbuild
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
    exiftool
  ];
}
