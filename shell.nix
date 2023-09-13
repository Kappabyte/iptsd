{ pkgs ? import <nixpkgs> {} }:
  pkgs.mkShell {
    # nativeBuildInputs is usually what you want -- tools you need to run
    nativeBuildInputs = [ pkgs.ninja pkgs.meson pkgs.cmake pkgs.pkg-config pkgs.python3 ];
}

