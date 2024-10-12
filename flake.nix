{
  description = "Container Orchestrator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      pkgs = import nixpkgs { system = "x86_64-linux"; };
    in
    {
      devShell.x86_64-linux = nixpkgs.legacyPackages.x86_64-linux.mkShell {
        buildInputs = with nixpkgs.legacyPackages.x86_64-linux; [
          go
          pkgs.linux-pam
          pkgs.pkg-config
        ];

        PAM_PATH = "${pkgs.linux-pam}/lib/pkgconfig";

        shellHook = ''
          echo "Setting up dev directories"
          mkdir -p logs

          export MOUNTS_PATH="/srv/mounts"
          sudo mkdir -p /srv/mounts

          if ! getent group sftpusers > /dev/null 2>&1; then
            echo "Creating sftpusers group"
            sudo groupadd sftpusers
          else
            echo "sftpusers group already exists"
          fi

          echo "MOUNTS_PATH=$MOUNTS_PATH"

          echo "Setting control node variables"
          export ETCD_NAMESPACE="gameservers"
          echo "ETCD_NAMESPACE=$ETCD_NAMESPACE"
          export CONTAINERD_NAMESPACE="gameservers"
          echo "CONTAINERD_NAMESPACE=$CONTAINERD_NAMESPACE"

          echo "Setting worker node variables"
          export NODE_ID="node-1"
          echo "NODE_ID=$NODE_ID"
          export CONTROL_NODE_URI="http://localhost:3001/api"
          echo "CONTROL_NODE_URI=$CONTROL_NODE_URI"
          export CONTAINERD_PATH="/run/containerd/containerd.sock"
          echo "CONTAINERD_PATH=$CONTAINERD_PATH"
          export LOGS_PATH="/home/kowalski/dev/server-hosting/container-orchestrator/logs"
          echo "LOGS_PATH"=$LOGS_PATH
        '';

      };
    };
}



