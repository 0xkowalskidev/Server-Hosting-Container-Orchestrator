{ config, lib, pkgs, ... }:


let
  jsonContent = builtins.toJSON {
    namespace= "development";
    nodeIp= "192.168.1.30";
    controlNodeIp= "192.168.1.30";
    containerdSocketPath="/run/containerd/containerd.sock";

    storagePath= "/home/admin/mounts/";

    cniPath= "/run/current-system/sw/bin";
    networkConfigPath="/etc/cni/net.d";
    networkConfigFileName="mynet";
    networkNamespacePath="/var/run/netns/";

    logPath= "/home/admin/logs/";
  };

  configFile = pkgs.writeText "config.json" jsonContent;
  
  workerNodeBinary = pkgs.copyPathToStore "/home/kowalski/dev/server-hosting/container-orchestrator/bin/worker-node";
in
{
  users.users.admin = {
    isNormalUser = true;
    extraGroups = [ "wheel" ]; # Wheel is sudo
    initialPassword = "pass"; # temp
        openssh.authorizedKeys.keys = [ "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCipJMhdmebJ1lOE2KJ3p2PAKktfvAPPtCZ0xx+u1SMsD8dhEY7LpHCwLA3nav5bGZViUEwIqihW8TyBCCZADUPzKLuJXsJ5Dd5QDOb0apFbklDw7LoamuCe2Lr89W0hwG5R+gSks1A6bz0nlet+X3xyrJXEHwB0Bq+06Fgv6jMxLN5I+U8ASfujuWXLh5WyLIhS6/yOmOOhFD60+zhTdEPlyllVyL1n4eYtogVSu9qshGMUvDyqDHlWOGZOE4RXmeecLoVNj9x3z9IXwVQ7TJbu+mBHXZwXPRUIWNjeha5HkSfytmrHNk+8iVdPfvArNrhmFoHJ7njXAZjAYq7s8OcqrqXoDGH3TKSRGIh/hLphpvRzYXkgNVL5F6/zISWfCEoP7mUFvmlthHS2jLeamxJLJ4/mT7+QfiBey5r5sso/r1MW+KzZAhlO/gB8bS9lU7cJk4eqo+ZAhLNnWNZnb1i+bB6QBS1wYplx4iCjX5Z/y38X4NEFoMdYRe95xcrsR3vI5PVVQthbzd8qFJeFHmjWkHGFTTi+7sXIDfwDfA+50XvzxlQHNhe4I/Y9zAsoizfv5Yu2HQvcammVskKjES6aF0HTe2QoFLDDuFX4wP1YfiZW7aT0OyJFL2bfqIFLvMM18PX+mtH5GHg/+14eodxR9JNlx0nsdp8G9dOIEg/kw== 0xkowalskiaudit@gmail.com"];
  };


  networking = {
    hostName = "worker-node";  # Set a hostname
  };

    # Set up a basic firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 8081 ];
    allowedTCPPortRanges = [ { from = 30000; to = 32767; } ];
  };

   services.openssh = {
    enable = true;
  };

    # Ensure the binary is executable and copied to the desired location within the VM
  systemd.services.copy-worker-node = {
    description = "Ensure worker-node binary is in place";
    after = [ "network.target" ];
    wantedBy = [ "multi-user.target" ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = "true";
    };
    script = ''
      mkdir -p /home/admin
      cp ${workerNodeBinary} /home/admin/worker-node
      chmod +x /home/admin/worker-node
    '';
  };

  systemd.services.setup-config-json = {
    description = "Setup config.json in /home/admin";
    after = [ "network.target" ];
    wantedBy = [ "multi-user.target" ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = "true";
    };
    script = ''
      mkdir -p /home/admin/mounts
      mkdir -p /home/admin/logs
      mkdir -p /home/admin
      cp ${configFile} /home/admin/config.json
    '';
  };

  environment.systemPackages = with pkgs; [
    cni
    cni-plugins
    containerd
  ];



virtualisation.containerd.enable = true;
  
 environment.etc."cni/net.d/10-mynet.conflist".text = ''
 {
  "cniVersion": "1.0.0",
  "name": "mynet",
  "plugins": [
    {
      "type": "bridge",
      "bridge": "cni0",
      "isGateway": true,
      "ipMasq": true,
      "ipam": {
        "type": "host-local",
        "subnet": "10.22.0.0/16",
        "routes": [
          { "dst": "0.0.0.0/0" }
        ]
      }
    },
    {
      "type": "portmap",
      "capabilities": {
        "portMappings": true
      },
      "snat": true
    }
  ]
}
  '';


  # DONT CHANGE ME
  system.stateVersion = "23.11"; 
}

