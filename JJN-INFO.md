# JJN-INFO: SHORTNER

## Alright, Listen Up Future Me

So, you built this thing called `Shortner`. It's a tiny, self-hosted URL shortener written in Go because, let's be real, long URLs are a crime against humanity. You made this because your inbox was turning into an unmanageable hellscape of tracking-infested URLs, and you figured, "Hey, I can fix this!"

Also, let's be honest—you just wanted an excuse to mess around with Go again.

## What This Project Does (In Case You Forget)
- **Shortens URLs** because nobody needs to see `utm_source=nightmare&utm_campaign=chaos`
- **Works with Nginx** because why expose a Go app directly?
- **Stores links in SQLite** because you're too lazy to set up Postgres for this
- **Auto-cleans expired links** because clutter is bad
- **Self-hosted** because you don’t trust third-party services (good job, paranoid self)

## The Stuff You Always Forget (But Need)

### .env Configuration (Don't Skip This!)
Before running anything, copy the example `.env` and tweak it:
```bash
cp .env.example .env
```
Then, edit it with your **own damn values**:
```ini
LOG_PATH=app.log
LOG_LEVEL=debug
CACHE_EXPIRY=30  # Cleanup after X days
SHORT_DOMAIN=https://my.short.link  # Should match Nginx setup
PORT=8889
```

## Running This Locally (When You Inevitably Need to Debug)
1. **Check if you even have Go installed:**
   ```bash
   go version
   ```
   If that fails, well… go install Go.

2. **Run it in dev mode:**
   ```bash
   go run main.go
   ```

3. **Or build the binary and run it:**
   ```bash
   go build -o shortener
   ./shortener
   ```

4. **Test it works:**
   ```bash
   curl "http://localhost:8889/shorten?url=https://example.com?utm_source=spam"
   ```
   Expected output:
   ```
   https://my.short.link/a1b2c3
   ```

5. **Check redirection:**
   ```bash
   curl -L "https://my.short.link/a1b2c3"
   ```
   If it doesn’t work, time to go debugging!

## Deploying This Beast With GoReleaser
Because manually building stuff is for suckers.

### **Build It Locally First**
```bash
goreleaser release --snapshot --clean
```

That should:
- Compile binaries for Linux, macOS, and Windows (because why not?)
- Zip and tar.gz them up neatly
- Generate checksums so you feel like a pro

### **Tag a Release (Because You Always Forget)**
```bash
git add .
git commit -m "Prepare release v1.0.0"
git tag -a v1.0.0 -m "First official release"
git push origin v1.0.0
```
Boom. That triggers **GitHub Actions**, builds everything, and uploads it automagically.

## Running It as a Systemd Service (So It Survives Reboots)
Create the service file `/etc/systemd/system/shortener.service`:
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

Reload systemd and start it:
```bash
sudo systemctl daemon-reload
sudo systemctl enable shortener
sudo systemctl start shortener
```

Check if it's alive:
```bash
sudo systemctl status shortener
```
If it’s dead, time to dig into logs and cry a little.

## Future Me’s To-Do List
- **Improve Logging** – Because right now it’s "just enough" and that’s never actually enough.
- **Better Expiry Handling** – Maybe let users choose different expiration rules?
- **Admin Dashboard?** – Maybe, if you're feeling fancy.
- **Track Usage Stats** – Because, why not? Would be cool to know how many links you shorten.

---
Okay, that's everything. If you're reading this, you either broke something or forgot how your own project works. Either way, good luck, future me. You're gonna need it. 🚀

