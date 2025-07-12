1. `sudo cp v1nage.service /etc/systemd/system/v1nage.service`
2. `sudo cp config.yml /etc/v1nage-config.yml`
3. `sudo cp v1nage /usr/local/bin/v1nage`
4. `sudo chmod +x /usr/local/bin/v1nage`
5. `sudo systemctl daemon-reload`
6. `sudo systemctl enable v1nage.service`
7. `sudo systemctl start v1nage.service`
8. `sudo systemctl status v1nage.service`
