<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./media/xui-im-dark.png">
    <img alt="xui-im" src="./media/xui-im-light.png">
  </picture>
</p>

[![Release](https://img.shields.io/github/v/release/codewithtamim/xui-im.svg)](https://github.com/codewithtamim/xui-im/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/codewithtamim/xui-im/release.yml.svg)](https://github.com/codewithtamim/xui-im/actions)
[![GO Version](https://img.shields.io/github/go-mod/go-version/codewithtamim/xui-im.svg)](#)
[![Downloads](https://img.shields.io/github/downloads/codewithtamim/xui-im/total.svg)](https://github.com/codewithtamim/xui-im/releases/latest)
[![License](https://img.shields.io/badge/license-GPL%20V3-blue.svg?longCache=true)](https://www.gnu.org/licenses/gpl-3.0.en.html)

> This is an enhanced fork of [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) with additional features such as API key management, Swagger API documentation, and more planned improvements.

XUI-IM is an advanced, open-source web-based control panel for managing Xray-core servers. It offers a user-friendly interface for configuring and monitoring various VPN and proxy protocols.

## What's Different in This Fork?

- API key management: create, manage, and revoke API keys from Panel Settings. Authenticate API requests with the `X-API-Key` header instead of session cookies.
- Swagger API documentation: built-in Swagger UI at `/panel/api/docs`, with a toggle to enable or disable it from the settings page.
- More features coming soon: this fork is under active development with additional enhancements planned.

## Quick Start

```bash
bash <(curl -Ls https://raw.githubusercontent.com/codewithtamim/xui-im/main/install.sh)
```

For full documentation, please visit the [project Wiki](https://github.com/codewithtamim/xui-im/wiki).

## Docker

```bash
docker pull c0dewithtamim/xui-im:latest
```

Or use docker-compose:

```bash
git clone https://github.com/codewithtamim/xui-im.git
cd xui-im
docker-compose up -d
```

## Credits

This project is a fork of [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Huge thanks to the original authors and contributors:

- [MHSanaei](https://github.com/MHSanaei), original creator and maintainer of 3X-UI
- [alireza0](https://github.com/alireza0/), major contributor to the original project

## Acknowledgment

- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (License: GPL-3.0): _Enhanced v2ray/xray and v2ray/xray-clients routing rules with built-in Iranian domains and a focus on security and adblocking._
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (License: GPL-3.0): _This repository contains automatically updated V2Ray routing rules based on data on blocked domains and addresses in Russia._

## License

This project is licensed under the [GPL-3.0 License](https://www.gnu.org/licenses/gpl-3.0.en.html).
