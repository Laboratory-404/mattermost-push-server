# Mattermost Push Proxy Server for Linux (Ubuntu)

[![Go Version](https://img.shields.io/badge/go-1.22%2B-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-linux--ubuntu-orange.svg)](https://ubuntu.com)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A high-performance, lightweight, self-hosted **Push Notification Proxy Server** for **Mattermost Dedicated Servers**. 

Designed to run alongside your Mattermost server on **Ubuntu Linux**, enabling instant push notifications for **Android (FCM)**, **iOS (APNs)**, **Mattermost Desktop Client**, and **Web Browser** versions without relying on third-party cloud push subscriptions.

---

## 🚀 Key Features

* **Native Ubuntu Linux Service**: Runs as a low-overhead `systemd` service (~15MB RAM footprint).
* **FCM HTTP v1 API**: Uses Google's latest Firebase HTTP v1 API with Service Account OAuth2 token caching.
* **APNs HTTP/2 Support**: Uses Apple Push Notification service HTTP/2 protocol supporting `.p8` JWT Token authentication and `.p12` certificates.
* **Cross-Platform Compatibility**: Supports Android, iOS (React Native & native builds), Mattermost Desktop App, and Web client notifications.
* **Built-in Rate Limiting & Security**: Protects your server with token-bucket rate limiting, security headers, and panic recovery middleware.
* **Health & Metrics Endpoint**: Real-time stats at `/metrics` and health checks at `/healthz`.
* **Testing CLI Utility**: Includes `test-push` CLI tool to verify notifications without making real channel posts.
* **Docker Support**: Includes multi-stage `Dockerfile` and `docker-compose.yml`.

---

## 🏗️ How It Works

```
┌────────────────────────────────────────────────────────┐
│               Ubuntu Linux Server                      │
│                                                        │
│  ┌──────────────────┐           ┌───────────────────┐  │
│  │ Mattermost Server│ HTTP POST │ Push Proxy Server │  │
│  │  (Port 8065)     ├──────────►│ (Port 8066)       │  │
│  └──────────────────┘           └─────────┬─────────┘  │
└───────────────────────────────────────────┼────────────┘
                                            │
               ┌────────────────────────────┼───────────────────────────┐
               │                            │                           │
               ▼                            ▼                           ▼
    ┌────────────────────┐       ┌────────────────────┐       ┌────────────────────┐
    │  Google FCM v1     │       │     Apple APNs     │       │   Web & Desktop    │
    └──────────┬─────────┘       └──────────┬─────────┘       └──────────┬─────────┘
               │                            │                            │
               ▼                            ▼                            ▼
      Android Devices                   iOS Devices              Desktop / Web Clients
```

1. When a mention or direct message occurs, **Mattermost Server** sends an HTTP POST payload to `http://127.0.0.1:8066/api/v1/send_push`.
2. The **Push Proxy Server** identifies the target device platform (`android_rn`, `apple_rn`, `web`, `desktop`).
3. The Proxy dispatches the payload to **FCM HTTP v1** (for Android) or **APNs HTTP/2** (for iOS).

---

## 📦 Quick Installation on Ubuntu Linux

### Option 1: Automated Installation Script (Recommended)

Run the automated installer on your Ubuntu server:

```bash
git clone https://github.com/Laboratory-404/mattermost-push-server.git
cd mattermost-push-server
sudo bash scripts/install.sh
```

The script will:
1. Create a dedicated `mattermost-push` system user.
2. Build/copy the binaries to `/usr/local/bin/`.
3. Create configuration directory `/etc/mattermost-push-proxy/`.
4. Install and enable the `systemd` service (`mattermost-push-proxy.service`).

---

### Option 2: Docker / Docker Compose

If you prefer using Docker on your Ubuntu server:

```bash
cd docker
sudo docker-compose up -d
```

---

## ⚙️ Configuration Setup

Edit `/etc/mattermost-push-proxy/config.yaml`:

```yaml
server:
  listen_address: "0.0.0.0:8066"
  enable_rate_limit: true
  rate_limit_per_sec: 100

log:
  level: "info"
  log_to_file: true
  file_path: "/var/log/mattermost-push-proxy.log"

fcm:
  enabled: true
  credentials_file: "/etc/mattermost-push-proxy/fcm-service-account.json"
  android_rn_topic: "com.mattermost.rn"

apns:
  enabled: true
  auth_key_file: "/etc/mattermost-push-proxy/AuthKey_APNS.p8"
  key_id: "YOUR_KEY_ID"
  team_id: "YOUR_TEAM_ID"
  topic: "com.mattermost.rn"
  production: true
```

### 1. Setting up FCM (Android)
1. Go to [Firebase Console](https://console.firebase.google.com/) -> Project Settings -> **Service Accounts**.
2. Click **Generate New Private Key** to download `service-account.json`.
3. Save the JSON file to `/etc/mattermost-push-proxy/fcm-service-account.json`.

### 2. Setting up APNs (iOS)
1. Go to [Apple Developer Portal](https://developer.apple.com/account/) -> Certificates, Identifiers & Profiles -> **Keys**.
2. Create an **Apple Push Notifications service (APNs)** key and download the `.p8` file.
3. Save the file to `/etc/mattermost-push-proxy/AuthKey_APNS.p8`.
4. Fill in `key_id`, `team_id`, and `topic` (App Bundle Identifier) in `config.yaml`.

---

## 🔧 Configuring Mattermost Server

1. Log into your Mattermost Server as an Admin.
2. Open **System Console** -> **Environment** -> **Push Notification Server** (or **Notifications** -> **Push Notification Server**).
3. Configure the following settings:
   * **Enable Push Notifications**: `True`
   * **Push Notification Server Type**: `Manually enter Push Notification Server location`
   * **Push Notification Server**: `http://127.0.0.1:8066`
   * **Push Notification Contents**: `Full message content in push notifications`
4. Click **Save**.

---

## 🧪 Testing Push Notifications

Run the test utility on your server to verify connectivity:

```bash
# Test Android Push
test-push -server http://localhost:8066 -platform android_rn -token YOUR_FCM_DEVICE_TOKEN -message "Hello from Push Proxy!"

# Test iOS Push
test-push -server http://localhost:8066 -platform apple_rn -token YOUR_APNS_DEVICE_TOKEN -message "Hello iOS!"
```

---

## 📊 Management & Monitoring

### Service Management Commands
```bash
# Start service
sudo systemctl start mattermost-push-proxy

# Check status
sudo systemctl status mattermost-push-proxy

# View live logs
tail -f /var/log/mattermost-push-proxy.log
```

### Health & Metrics Endpoints
* **Health Check**: `curl http://localhost:8066/healthz`
* **Stats / Metrics**: `curl http://localhost:8066/metrics`

---

## 📁 Repository Structure

```
.
├── cmd/
│   ├── server/           # Proxy server entry point (main.go)
│   └── test-push/        # Test CLI tool
├── pkg/
│   ├── config/           # YAML & Environment config loader
│   ├── server/           # HTTP server, handlers & security middleware
│   ├── push/             # Notification service dispatcher
│   │   ├── fcm/          # Firebase HTTP v1 client
│   │   ├── apns/         # APNs HTTP/2 client
│   │   └── web/          # Web/Desktop notification bridge
│   └── metrics/          # Monitoring & statistics engine
├── config/
│   ├── config.example.yaml
│   ├── mattermost-config.json
│   └── nginx.conf
├── scripts/
│   ├── install.sh        # One-line Ubuntu installer
│   ├── build.sh          # Cross-compilation script
│   └── mattermost-push-proxy.service
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── Makefile
└── README.md
```

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for details.
