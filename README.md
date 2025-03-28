# v1nage

[![goreleaser](https://github.com/nsupc/v1nage/actions/workflows/release.yml/badge.svg)](https://github.com/nsupc/v1nage/actions/workflows/release.yml)

Named in honor of Europeian Honoured Citizen and Former Delegate [Vinage Vinage](https://www.nationstates.net/page=boneyard?nation=delegate_vinage), v1nage is designed to send telegrams to new World Assembly members in Europeia and alert the Regional Security Council to their presence via a Discord webhook.

## Installation (Linux CLI)

1. `mkdir v1nage && cd v1nage`: create a directory to store project files and enter it
2. `wget https://github.com/nsupc/v1nage/releases/download/0.2.1/v1nage_0.2.1_linux_amd64.tar.gz`: download the v1nage binary. Note that version 0.2.1 was the latest release at the time of writing, please see the [releases](https://github.com/nsupc/v1nage/releases) page for the most updated version of the tool -- you can just copy the URL for the amd64 release.
3. `tar -xzf v1nage_0.2.1_linux_amd64.tar.gz`: unzip the archive. Again, be sure to use the most recent version number.
4. `rm v1nage_0.2.1_linux_amd64.tar.gz`: [optional] delete the archive.
5. `chmod 500 ./v1nage`: set the binary executable bit.
6. `nohup ./v1nage [arguments] &`: run the tool disconnected from your terminal session.
7. `ps -aux`: [optional] ensure that the program is running.

## CLI Flags

- --region [string]: region to subscribe to events from
- --url [string]: base URL for eurocore instance (for sending telegrams)
- --user [string]: eurocore username
- --password [string]: eurocore user password
- --webhook-id [string]: Discord webhook id
- --webtook-token [string]: Discord webhook token
- --telegram-sender [string]: nation sending telegrams
- --new-wa-telegram-id [string]: NS telegram ID for telegram to be sent to new WA joins
- --new-wa-telegram-secret [string]: NS telegram secret for telegram to be sent to new WA joins
- --move-telegram-id [string]: NS telegram ID for telegram to be sent to WA nations that move to region
- --move-telegram-secret [string]: NS telegram secret for telegram to be sent to WA nations that move to region
