This repo is meant to complement the tutorial written for the 2017 GopherAcademy Advent series.

Users running this tutorial will need a linux host as the plugins depend on `systemd`. 
You will also need to have osquery installed on the system. https://osquery.io/downloads/

To see the Twitter distributed plugin in action, first [create a Twitter app](https://apps.twitter.com/app/new) and then update the environment variables in `env`.
Once you have the right environment variables defined, use `source env` to add them to your shell. 

Run `make osqueryi` to start an osquery console you can run interactively for the `systemd` table. 
Run `make osqueryd` to run osqueryd as a daemon. This will give you acccess to the config and logger plugins.
