#!/usr/bin/env bash
# Automated Installation Script for Mattermost Push Proxy on Ubuntu Linux
set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}====================================================${NC}"
echo -e "${BLUE} Installing Mattermost Push Proxy Server on Ubuntu ${NC}"
echo -e "${BLUE}====================================================${NC}"

if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Please run as root (e.g., sudo ./scripts/install.sh)${NC}"
  exit 1
fi

CONFIG_DIR="/etc/mattermost-push-proxy"
LOG_DIR="/var/log"
BINARY_DEST="/usr/local/bin/mattermost-push-proxy"
TEST_BINARY_DEST="/usr/local/bin/test-push"

echo -e "${GREEN}1. Creating system user 'mattermost-push'...${NC}"
id -u mattermost-push &>/dev/null || useradd -r -s /bin/false mattermost-push

echo -e "${GREEN}2. Creating configuration directories at ${CONFIG_DIR}...${NC}"
mkdir -p "$CONFIG_DIR"
chown -R mattermost-push:mattermost-push "$CONFIG_DIR"
chmod 750 "$CONFIG_DIR"

echo -e "${GREEN}3. Building Go binaries...${NC}"
if command -v go &> /dev/null; then
  go build -ldflags="-s -w" -o build/mattermost-push-proxy ./cmd/server
  go build -ldflags="-s -w" -o build/test-push ./cmd/test-push
  cp build/mattermost-push-proxy "$BINARY_DEST"
  cp build/test-push "$TEST_BINARY_DEST"
else
  echo -e "${RED}Go compiler not found. Attempting to copy pre-compiled binary...${NC}"
  if [ -f "./build/mattermost-push-proxy" ]; then
    cp ./build/mattermost-push-proxy "$BINARY_DEST"
    cp ./build/test-push "$TEST_BINARY_DEST"
  else
    echo -e "${RED}Please install Go (apt install golang-go) or compile the binary first.${NC}"
    exit 1
  fi
fi

chmod +x "$BINARY_DEST" "$TEST_BINARY_DEST"

echo -e "${GREEN}4. Setting up configuration file...${NC}"
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
  cp config/config.example.yaml "$CONFIG_DIR/config.yaml"
  chown mattermost-push:mattermost-push "$CONFIG_DIR/config.yaml"
  chmod 600 "$CONFIG_DIR/config.yaml"
  echo -e "${GREEN}Created default configuration file at $CONFIG_DIR/config.yaml${NC}"
else
  echo -e "${BLUE}Existing config file preserved at $CONFIG_DIR/config.yaml${NC}"
fi

echo -e "${GREEN}5. Setting up log file...${NC}"
touch "$LOG_DIR/mattermost-push-proxy.log"
chown mattermost-push:mattermost-push "$LOG_DIR/mattermost-push-proxy.log"
chmod 644 "$LOG_DIR/mattermost-push-proxy.log"

echo -e "${GREEN}6. Installing Systemd Service...${NC}"
cp scripts/mattermost-push-proxy.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable mattermost-push-proxy.service

echo -e "${BLUE}====================================================${NC}"
echo -e "${GREEN}✅ Installation Completed Successfully!${NC}"
echo -e "${BLUE}====================================================${NC}"
echo -e "Next steps:"
echo -e "1. Edit configuration file:"
echo -e "   ${BLUE}sudo nano /etc/mattermost-push-proxy/config.yaml${NC}"
echo -e "2. Add your Firebase Service Account JSON / Apple .p8 keys into ${CONFIG_DIR}/"
echo -e "3. Start the service:"
echo -e "   ${BLUE}sudo systemctl start mattermost-push-proxy${NC}"
echo -e "4. Check service status & logs:"
echo -e "   ${BLUE}sudo systemctl status mattermost-push-proxy${NC}"
echo -e "   ${BLUE}tail -f /var/log/mattermost-push-proxy.log${NC}"
echo -e "5. Update Mattermost Server System Console -> Push Notification Server:"
echo -e "   ${BLUE}http://127.0.0.1:8066${NC}"
