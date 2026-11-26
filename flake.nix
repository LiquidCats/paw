{
  description = "paw services, built reproducibly with Nix";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { nixpkgs, ... }:
    let
      inherit (nixpkgs) lib;
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      # every directory in ./cmd is a buildable service: nix build .#rater
      services = lib.attrNames (lib.filterAttrs (_: type: type == "directory") (builtins.readDir ./cmd));

      src = lib.fileset.toSource {
        root = ./.;
        fileset = lib.fileset.unions [ ./go.mod ./go.sum ./cmd ./internal ./pkg ./tests ];
      };
    in
    {
      packages = lib.genAttrs systems (
        system:
        lib.genAttrs services (
          name:
          nixpkgs.legacyPackages.${system}.buildGo127Module {
            pname = name;
            version = "dev";
            inherit src;
            # update after go.mod/go.sum changes: set to lib.fakeHash, build, copy the hash from the error
            vendorHash = "sha256-JdvMMCK06kiuxVDUyem10c1jAH0THd2WYQnXDlK6u1o=";
            subPackages = [ "cmd/${name}" ];
            env.CGO_ENABLED = 0;
            ldflags = [ "-s" "-w" ];
          }
        )
      );
    };
}
