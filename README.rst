Web Alert Indicator
===================

`wali` is a web alert indicator, i.e., websites will be regularly checked for
changes and on change notifications will be sent via email.

**DISCLAIMER**: This tool is at an early development stage and not fully tested
                yet, please do not use it in production.


Build
-----

Ensure that the git directory is cleanly committed and a tag is set.

Download the docker image of goreleaser-cross that enables cross compiling:

::

    sudo docker pull goreleaser/goreleaser-cross

Build current snapshot (without release tags) locally, including deb:

::

    sudo docker run --rm --privileged -v $(pwd):$(pwd) \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -w $(pwd) goreleaser/goreleaser-cross release --clean


Afterwards all created binaries and package can be found in `dist/` directory.


Install
-------

Install the package on Debian:

::

    sudo dpkg -i wali_<version>_linux_amd64.deb

Adapt the config to your needs `/etc/wali.yaml`.


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

or if installed, just

::

    wali /etc/wali.yaml

On start-up the settings of the YAML file will be transferred to the wali
sqlite database. In intervals of 1 s the database will be checked by the
scheduler if a webpage check needs to be triggered based on last checked
timestamp from the database and the frequency. If a timestamp of a webpage
exceeded, it will be downloaded and hashed, if the hash differs from the
stored one in the database, an email will be sent using the SMTP information
to inform about the change.
