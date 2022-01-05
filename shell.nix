with import <nixpkgs> {};

stdenv.mkDerivation {
    name = "rye";
    buildInputs = [
      go_1_17
      goimports
    ];
    shellHook = ''
      export GO111MODULE=auto
      export GOROOT="${pkgs.go_1_17}/share/go"
      export PATH=$PWD/bin:$PATH
    '';
}
