rm control-node.qcow2
nix-build '<nixpkgs/nixos>' -A vm \
-I nixpkgs=channel:nixos-23.11 \
-I nixos-config=./configuration.nix

