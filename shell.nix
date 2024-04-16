{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [ pkgs.go ];

  packages = with pkgs; [
    cope #cmake
  ];

}
