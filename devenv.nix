{
  pkgs,
  lib,
  config,
  ...
}:
{
  # https://devenv.sh/languages/
  languages.go.enable = true;
  languages.javascript = {
    enable = true;
    yarn = {
      enable = true;
      install.enable = true;
    };
  };

  # https://devenv.sh/services/
  services.redis.enable = true;
  services.mongodb.enable = true;

  # packages = with pkgs;  [];


  # See full reference at https://devenv.sh/reference/options/
}
