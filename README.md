# v2rayA

[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/v2rayA/v2raya)](https://hub.docker.com/r/mzz2017/v2raya)
[![Travis (.org)](https://img.shields.io/travis/v2rayA/v2rayA?label=travis-ci%20build)](https://travis-ci.org/v2rayA/v2rayA)
[![License: AGPL v3-only](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

[**English**](https://github.com/v2rayA/v2rayA/blob/main/README.md) · [**简体中文**](https://github.com/v2rayA/v2rayA/blob/main/README_zh.md)

---

## About v2rayA

v2rayA is a V2Ray client supporting global transparent proxy on Linux and system proxy on Windows and macOS. It is compatible with SS, SSR, Trojan (trojan-go), Tuic, and [Juicity](https://github.com/juicity) protocols. [SSR protocol list](https://github.com/v2rayA/shadowsocksR/blob/main/README.md#ss-encrypting-algorithm).

We are committed to providing the simplest operation and meeting most needs. Thanks to the advantages of the Web GUI, you can use it on your local computer or easily deploy it on a router or NAS.

## Features

- **Multi-protocol support:** VMess, VLESS, Shadowsocks, ShadowsocksR, Trojan, TUIC, Juicity
- **Global transparent proxy:** TProxy, TUN, and system proxy support
- **Web-based GUI:** Manage your proxy settings via any modern browser
- **Cross-platform:** Runs on Linux, Windows, and macOS
- **Automatic Updates:** Auto-update PAC files, GFW lists, and subscriptions
- **Routing rules:** Powerful routing capabilities based on RoutingA and V2Ray/XRay routing
- **Containerized:** Official Docker images available

## Architecture

Please refer to [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architectural information.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Specification](docs/SPECIFICATION.md)
- [Contributing](docs/CONTRIBUTING.md)
- [Roadmap](docs/ROADMAP.md)
- [Official Documentation](https://v2raya.org/en/docs/prologue/introduction/)

## Installation

v2rayA provides the following methods of installation:

1. **Package managers:**
   - Debian/Ubuntu: APT Repository
   - Arch Linux: AUR (`v2raya`)
   - macOS: Homebrew (`v2rayA/v2raya/v2raya`)
   - Windows: Scoop (`v2raya`), Winget (`v2rayA.v2rayA`)
   - Ubuntu Snap: `snap install v2raya`
2. **Docker:** [Official Docker Hub](https://hub.docker.com/r/v2rayA/v2raya)
3. **OpenWrt:** [v2raya-openwrt](https://github.com/v2rayA/v2raya-openwrt) and official repos (from OpenWrt 22.03)
4. **Binary/Installer:** Download from [GitHub Releases](https://github.com/v2rayA/v2rayA/releases)

See [Installation Guide](https://v2raya.org/en/docs/prologue/introduction/) for detailed instructions.

## Quick Start

### 1. Install

Follow the installation steps above for your platform.

### 2. Access Web Interface

By default, v2rayA listens on `http://127.0.0.1:2017`. Open this URL in your browser.

### 3. Import Servers

Paste your subscription link or import server configurations directly.

### 4. Connect

Select a server and click "Connect". Enable "Transparent Proxy" in settings for system-wide proxying.

## Configuration

v2rayA configuration is stored in `~/.config/v2raya/` (Linux) or `~/.v2raya/` (macOS/Windows) by default. You can specify a custom config directory using the `--config` flag or `V2RAYA_CONFIG_DIR` environment variable.

## Statement

1. The program does not store any user data in the cloud. All user data is stored locally.
2. **Do not use this project for illegal purposes.**

## Credits

- [hq450/fancyss](https://github.com/hq450/fancyss)
- [ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)
- [nadoo/glider](https://github.com/nadoo/glider)
- [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)
- [zfl9/ss-tproxy](https://github.com/zfl9/ss-tproxy/blob/master/ss-tproxy)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/v2rayA/v2rayA.svg)](https://starchart.cc/v2rayA/v2rayA)
