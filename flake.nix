{
  description = "Briggette - A tool for monitoring of the Optimism Bridge";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";

    systems.url = "github:nix-systems/default";

  };

  outputs = { self, nixpkgs, systems, ... }@inputs:
    let
      eachSystem = f:
        nixpkgs.lib.genAttrs (import systems)
        (system: f system nixpkgs.legacyPackages.${system});
    in {

      devShells = eachSystem (system: pkgs: {
        default = pkgs.mkShell {
          shellHook = ''
            # Set here the env vars you want to be available in the shell
          '';
          hardeningDisable = [ "all" ];

          packages = with pkgs; [ go sqlite sqlc okteto kubectl kubelogin-oidc ];
        };
      });
    };
}
