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
        export NAMESPACE="gameservers"
        export CONTAINERD_PATH="/run/containerd/containerd.sock"
        echo "Environment variables set for development:"
        echo "NAMESPACE=$NAMESPACE"
        echo "CONTAINERD_PATH=$CONTAINERD_PATH"
      '';

    };
  };
}



