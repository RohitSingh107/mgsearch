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
  services = {
    redis.enable = true;
    mongodb = {
      enable = true;
      # initDatabaseUsername = "user";
      # initDatabasePassword = "password";
    };
  };

  env = {
    # DATABASE_URL = "mongodb://${config.services.mongodb.initDatabaseUsername}:${config.services.mongodb.initDatabasePassword}@127.0.0.1:27017/mgsearch";
    DATABASE_URL = "mongodb://127.0.0.1:27017/mgsearch";
  };
  #packages = with pkgs;  [];


  # See full reference at https://devenv.sh/reference/options/
}
