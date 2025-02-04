# theGeekist/Shortner  

Shortner is a super simple, self-hosted URL shortener written in Go that works seamlessly with Nginx. It’s designed for internal use cases where email links get excessively long and messy. This tool provides a clean, lightweight solution for quickly shortening URLs, making them easier to manage and share.  

💡 **Read more about the use case in our blog post:**  
👉 [Gmail Inbox Becoming Unmanageable? Here's One Fix With n8n and Telegram](https://geekist.co/gmail-inbox-becoming-unmanageable-heres-one-fix-with-n8n-and-telegram)  

## Features  

- **Simple & Lightweight:**  
  Minimal code with an HTTP API for shortening URLs.  
- **Custom Short Domain:**  
  Define your own short domain with `SHORT_DOMAIN`, e.g., `https://my.short.link/xyz123`.  
- **SQLite Database:**  
  Stores URL mappings in a self-contained SQLite DB.  
- **Automatic Cleanup:**  
  Removes links older than a configurable number of days (`CACHE_EXPIRY`).  
- **Environment-Driven Configuration:**  
  All settings (e.g., log path, log level, cache expiry) are managed via a `.env` file.  

---

## Installation & Setup  

### Environment Configuration  

Copy the example config file and update values as needed:  

```bash
cp .env.example .env
```  

Example `.env` file:  

```ini
LOG_PATH=app.log
LOG_LEVEL=debug
CACHE_EXPIRY=30  # days
SHORT_DOMAIN=https://my.short.link # should be configured on Nginx
PORT=8889
```  

---

### Installing Go (Ubuntu)  

Ensure your system has **Go 1.18+** installed:  

```bash
go version
```  

If your Go version is outdated, check the available version:  

```bash
apt-cache policy golang
```  

If needed, manually install the latest Go version:  

```bash
GO_VERSION=1.23.5  # Replace with latest release
wget https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version
```  

---

## Usage  

Shortner provides a simple HTTP API to shorten URLs and retrieve them later.  

### 1. Shorten a URL  

Send a GET request to `/shorten` with the `url` parameter:  

```bash
curl "http://localhost:8889/shorten?url=https://example.com/long-url?utm_source=spam&utm_campaign=clutter"
```  

Example response:  

```plaintext
https://my.short.link/a1b2c3
```  

This returns a **shortened version** of the URL using the `SHORT_DOMAIN` environment variable. Tracking parameters such as `utm_source` and `utm_campaign` are automatically removed.  

### 2. Redirect to the Original URL  

Accessing the shortened URL redirects you to the original destination:  

```bash
curl -L "https://my.short.link/a1b2c3"
```  

If the short code exists, it responds with an HTTP **302 Found** status and redirects to the cleaned URL.  

### 3. Automatic Cleanup  

Links older than `CACHE_EXPIRY` days (default **30 days**) are automatically deleted in the background.  

### 4. Running Behind Nginx  

To proxy requests through Nginx, add a configuration like this:  

```nginx
server {
    listen 80;
    server_name my.short.link;

    location / {
        proxy_pass http://localhost:8889;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```  

This allows users to access `https://my.short.link/xyz123` without needing to expose the Go service directly.

---

## Running as a Systemd Service  

To run Shortner as a background service on Ubuntu, create a systemd unit file at `/etc/systemd/system/shortener.service`:  

```ini
[Unit]
Description=Shortner Service
After=network.target

[Service]
ExecStart=/path/to/shortener
WorkingDirectory=/path/to
Restart=always
User=ubuntu
EnvironmentFile=/path/to/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```  

Reload systemd and start the service:  

```bash
sudo systemctl daemon-reload
sudo systemctl enable shortener
sudo systemctl start shortener
```  

Check its status:  

```bash
sudo systemctl status shortener
```  

---

## Contributions  

Contributions are welcome! If you'd like to add features, fix bugs, or improve documentation, feel free to submit a pull request.  

1. Fork the repository.  
2. Create a feature branch: `git checkout -b my-new-feature`.  
3. Commit your changes: `git commit -am 'Add some feature'`.  
4. Push to the branch: `git push origin my-new-feature`.  
5. Open a pull request.  

For major changes, please open an issue first to discuss your proposal.  

---

## License  

This project is licensed under the **MIT License**. See the [`LICENSE`](LICENSE) file for details.  
