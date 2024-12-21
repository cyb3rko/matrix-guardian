# 🛡️ Matrix Guardian 🛡️
[![github](https://img.shields.io/github/release/cyb3rko/matrix-guardian.svg?logo=github)](https://github.com/cyb3rko/matrix-guardian/releases/latest)
[![last commit](https://img.shields.io/github/last-commit/cyb3rko/matrix-guardian?color=FE5196&logo=git&logoColor=white)](https://github.com/cyb3rko/matrix-guardian/commits/main)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)
[![license](https://img.shields.io/github/license/cyb3rko/matrix-guardian?color=1BCC1B&logo=apache)](https://www.mozilla.org/en-US/MPL/2.0/)

- [Features](#features)
  - [URL Filter 🌐](#url-filter-)
  - [URL Phishing Check 🗡️](#url-phishing-check-)
    - [VirusTotal](#virustotal)
    - [FishFish](#fishfish)
a  - [File MIME Type Filter 📎](#file-mime-type-filter-)
  - [planned] *File Virus Scan* 🦠
  - [planned] *Keyword Filter* 📄
- [Protected Public Rooms (Mentions)](#protected-public-rooms-mentions)
- [License](#license)

---

## Features

### URL Filter 🌐

**Activation (default: true)**: `GUARDIAN_URL_FILTER: true|false`  
**Help Command**: `!gd url`

Guardian supports URL filtering based on a customizable domain list.

**Examples**:
- `!gd url block t.me`
- `!gd url unblock t.me`

### URL Phishing Check 🗡

Guardian supports checking URLs in messages for suspicious content.  
The analysis can be powered by the following providers:

#### VirusTotal

**Reference**: https://docs.virustotal.com/reference/url-info  
**API-Key (required)**: `GUARDIAN_VIRUS_TOTAL_KEY: <key>`  
**Activation (default: false)**: `GUARDIAN_URL_CHECK_VIRUS_TOTAL: true|false`

VirusTotal allows scanning a full URL and returning a very comprehensive scan report.  
Guardian rates a URL "suspicious" if the statistics `malicious` and `suspicious` have a combined score of 3 or more.

#### FishFish

**Reference**: https://fishfish.gg  
**Activation (default: false)**: `GUARDIAN_URL_CHECK_FISHFISH: true|false`

FishFish allows scanning a domain and returning a rating, if found in their reports.  
Guardian rates a URL "suspicious" if the FishFish rating is `malware` or `phishing` rather than `safe`.

### File MIME Type Filter 📎

**Activation (default: true)**: `GUARDIAN_MIME_FILTER: true|false`  
**Help Command**: `!gd mime`

Guardian supports file MIME type filtering based on a customizable MIME type list.

**Examples**:
- `!gd mime block application/zip`
- `!gd mime unblock application/zip`
- `!gd mime list`

## Protected Public Rooms (Mentions)

This list showcases some of the rooms who use the Matrix Guardian 🛡️:  
*If you would like to add a room, please open an [issue](https://github.com/cyb3rko/matrix-guardian/issues)*

## License

    This Source Code Form is subject to the terms of the Mozilla Public
    License, v. 2.0. If a copy of the MPL was not distributed with this
    file, You can obtain one at https://mozilla.org/MPL/2.0/.
