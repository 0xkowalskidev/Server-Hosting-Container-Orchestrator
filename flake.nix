{
  description = "Container Orchestrator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.legacyPackages.x86_64-linux.mkShell {
      buildInputs = with nixpkgs.legacyPackages.x86_64-linux; [
        go
      ];

      shellHook = ''
        export NAMESPACE_MAIN="gameservers"
        export RUNTIME_TYPE="containerd"
        export CONTAINERD_PATH="/run/containerd/containerd.sock"
        echo "Environment variables set for development:"
        echo "NAMESPACE_MAIN=$NAMESPACE_MAIN"
        echo "RUNTIME_TYPE=$RUNTIME_TYPE"
        echo "CONTAINERD_PATH=$CONTAINERD_PATH"
      '';

    };
  };
}



