# Configuring systemd services

You will need to configure the systemd services on the HCI appliance.

## ngrok

Make changes to the `/home/gigabyte/.ngrok2/ngrok.yml` to match needed settings.

## tinygohci

Use `sudo systemctl edit tinygohci` to edit the override settings for the web service:

```
[Service]
Environment="CITOKEN=putyourrealtokenhere"
Environment="GHWEBHOOKPATH=putyourrealhookhere"
Environment="CIWEBHOOKPATH=putyourrealhookhere"
Environment="GHORG=putyourrealorghere"
Environment="GHREPO=putyourrealrepohere"
Environment="GHKEY=putyourrealkeyhere"
Environment="GHKEYFILE=putyourrealkeyfilenamehere"
Environment="GHAPPID=putyourrealappidhere"
Environment="GHINSTALLID=putyourrealinstallidhere"
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/usr/local/tinygo/bin"
```
