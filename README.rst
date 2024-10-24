Web Alert Indicator
===================

`wali` is a web alert indicator, i.e., websites will be regularly checked for
changes and on change notifications will be sent via email.

DISCLAIMER: This tool is at an early development stage and not fully tested
            yet, please do not use it in production.


Build
-----

MacOS
~~~~~~

Ensure that goreleaser and nfpm are installed:

::
    brew install goreleaser nfpm dpkg


Build current snapshot (without release tags) locally, including deb:

::
    goreleaser release --snapshot --clean


Install
-------

Debian
~~~~~~

Install the package
::
    sudo dpkg -i wali_<version>_linux_386.deb

Adapt the config to your needs `/etc/vali.yaml`.


Start background systemd service:

::
    # reload systemd services
    sudo systemctl daemon-reload

    # start the service
    sudo systemctl start wali

    # check if running
    sudo systemctl status wali

    # emable the service on each start-up
    sudo systemctl enable wali


Usage
-----

Create `wali.yaml` and add corresponding webpages you want to watch, e.g.,

::
    smtp:
      host: mail.gmx.net
      port: 587
      username: <username>
      password: <password>
      from: <from_mail>
      to: <to_mail>
    webjobs:
      yourpage.de:
        url: https://www.yourpage.de
        xpath: /html/body/div/p
        frequency: 15m

Start `wali` to watch the pages defined in the YAML file:

::
    go run cmd/main.go run

On start-up the settings of the YAML file will be transferred to the wali
sqlite database. In intervals of 1 s the database will be checked by the
scheduler if a webpage check needs to be triggered based on last checked
timestamp from the database and the frequency. If a timestamp of a webpage
exceeded, it will be downloaded and hashed, if the hash differs from the
stored one in the database, an email will be sent using the SMTP information
to inform about the change.
