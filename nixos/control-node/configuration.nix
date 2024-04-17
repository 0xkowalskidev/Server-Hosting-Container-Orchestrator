{ config, lib, pkgs, ... }:


let
  jsonContent = builtins.toJSON {
    namespace= "development";
    nodeIp= "";
    controlNodeIp= "";
    containerdSocketPath="";

    storagePath= "";

    cniPath= "";
    networkConfigPath="";
    networkConfigFileName="";
    networkNamespacePath="";

    logPath= "";
  };

  configFile = pkgs.writeText "config.json" jsonContent;
  
  controlNodeBinary = pkgs.copyPathToStore "/home/kowalski/dev/server-hosting/container-orchestrator/bin/control-node";
in
{
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  users.users.admin = {
    isNormalUser = true;
    extraGroups = [ "wheel" ]; # Wheel is sudo
    initialPassword = "pass"; # temp
        openssh.authorizedKeys.keys = [ "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCipJMhdmebJ1lOE2KJ3p2PAKktfvAPPtCZ0xx+u1SMsD8dhEY7LpHCwLA3nav5bGZViUEwIqihW8TyBCCZADUPzKLuJXsJ5Dd5QDOb0apFbklDw7LoamuCe2Lr89W0hwG5R+gSks1A6bz0nlet+X3xyrJXEHwB0Bq+06Fgv6jMxLN5I+U8ASfujuWXLh5WyLIhS6/yOmOOhFD60+zhTdEPlyllVyL1n4eYtogVSu9qshGMUvDyqDHlWOGZOE4RXmeecLoVNj9x3z9IXwVQ7TJbu+mBHXZwXPRUIWNjeha5HkSfytmrHNk+8iVdPfvArNrhmFoHJ7njXAZjAYq7s8OcqrqXoDGH3TKSRGIh/hLphpvRzYXkgNVL5F6/zISWfCEoP7mUFvmlthHS2jLeamxJLJ4/mT7+QfiBey5r5sso/r1MW+KzZAhlO/gB8bS9lU7cJk4eqo+ZAhLNnWNZnb1i+bB6QBS1wYplx4iCjX5Z/y38X4NEFoMdYRe95xcrsR3vI5PVVQthbzd8qFJeFHmjWkHGFTTi+7sXIDfwDfA+50XvzxlQHNhe4I/Y9zAsoizfv5Yu2HQvcammVskKjES6aF0HTe2QoFLDDuFX4wP1YfiZW7aT0OyJFL2bfqIFLvMM18PX+mtH5GHg/+14eodxR9JNlx0nsdp8G9dOIEg/kw== 0xkowalskiaudit@gmail.com"];
  };


  networking = {
    hostName = "control-node";  # Set a hostname
  };

    # Set up a basic firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 2379 2380 8080 ]; 
  };

   services.openssh = {
    enable = true;
  };

 services.etcd = {
    enable = true;  # Enable etcd service
    # Set etcd to listen on all interfaces for client communication
    listenClientUrls = [ "http://0.0.0.0:2379" ];
    advertiseClientUrls = [ "http://localhost:2379" ];
    # For single-node, these can be the same as client URLs
    listenPeerUrls = [ "http://0.0.0.0:2380" ];
    initialAdvertisePeerUrls = [ "http://localhost:2380" ];
    # Set data directory
    dataDir = "/var/lib/etcd";
    # Single node setup; no need for discovery
        initialCluster = [ "default=http://localhost:2380" ];
    initialClusterState = "new";
    initialClusterToken = "etcd-cluster-1";
  };

    # Ensure the binary is executable and copied to the desired location within the VM
  systemd.services.copy-control-node = {
    description = "Ensure control-node binary is in place";
    after = [ "network.target" ];
    wantedBy = [ "multi-user.target" ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = "true";
    };
    script = ''
      mkdir -p /home/admin
      cp ${controlNodeBinary} /home/admin/control-node
      chmod +x /home/admin/control-node
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
      mkdir -p /home/admin
      cp ${configFile} /home/admin/config.json
    '';
  };
  
  # DONT CHANGE ME
  system.stateVersion = "23.11"; 
}

