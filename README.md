# v1nage

[![goreleaser](https://github.com/nsupc/v1nage/actions/workflows/release.yml/badge.svg)](https://github.com/nsupc/v1nage/actions/workflows/release.yml)

Named in honor of Europeian Honoured Citizen and Former Delegate [Vinage Vinage](https://www.nationstates.net/page=boneyard?nation=delegate_vinage), v1nage is designed to send telegrams to new World Assembly members in Europeia and alert the Regional Security Council to their presence via a Discord webhook.

## Installation (Linux CLI)

1. `mkdir v1nage && cd v1nage`: create a directory to store project files and enter it
2. `wget https://github.com/nsupc/v1nage/releases/download/0.3.3/v1nage_0.3.3_linux_amd64.tar.gz`: download the v1nage binary. Note that version 0.3.1 was the latest release at the time of writing, please see the [releases](https://github.com/nsupc/v1nage/releases) page for the most updated version of the tool -- you can just copy the URL for the amd64 release.
3. `tar -xzf v1nage_0.3.3_linux_amd64.tar.gz`: unzip the archive. Again, be sure to use the most recent version number.
4. `rm v1nage_0.3.3_linux_amd64.tar.gz`: [optional] delete the archive.
5. `mv config.yml.template config.yml`: if setting up for the first time, copy the config template to the config file. See [config](#config) for configuration options.
6. `nano config.yml`: open the config file in a text editor. Fill in the values and save the file.
7. `chmod 500 ./v1nage`: set the binary executable bit.
8. `nohup ./v1nage [arguments] &`: run the tool disconnected from your terminal session. See [arguments](#arguments) for available arguments.
9. `ps -aux`: [optional] ensure that the program is running.

## Config

- user: NationStates User-Agent
- region: region to subscribe to events from
- limit: maximum number of requests to send to NationStates in a 30 second period. 0 < limit < 50
- eurocore:
  - url: base URL for eurocore instance
  - username: eurocore username
  - password: eurocore user password
- webhook:
  - id: Discord webhook id
  - token: Discord webhook token
- join-message: message to post to webhook when a nation joins the WA in `region`. Use `$nation` in place of link to nation
- move-message: message to post to webhook when a WA nation moves to `region`. Use `$nation` in place of link to nation
- move-telegram:
  - id: NationStates telegram id for telegram sent to WA nations moving to `region`
  - secret: NationStates telegram secret for telegram sent to WA nations moving to `region`
  - author: NationStates nation responsible for sending telegram to WA nations moving to `region`.
- join-telegram:
  - id: NationStates telegram id for telegram sent to new WA joins in `region`
  - secret: NationStates telegram secret for telegram sent to new WA joins in `region`
  - author: NationStates nation responsible for sending telegram to new WA joins in `region`.
- log:
  - level: log level. ["DEBUG", "INFO", "WARN", "ERROR"]
  - token: BetterStack authorization token
  - endpoint: BetterStack logging endpoint

## Arguments

- [file]: optional path to configuration file. Defaults to ./config.yml if not set
