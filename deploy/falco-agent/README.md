# Falco Agent Installer

This directory contains a minimal first-phase installer for host-side Falco Agent.

## Files

- `install-falco-agent.sh`: installs the platform-side Falco Agent service on the client host
- `falco-agent.sh`: the runtime process of that Agent; it is not a separate Falco installer

## What It Does

- `install-falco-agent.sh` installs `falco-agent.service`, registers the host to the platform, and writes local config
- `falco-agent.sh` keeps heartbeats, reports status, pulls tasks, and reports task results
- Actual Falco package install/upgrade is triggered later by platform tasks such as `falco.install` and `falco.upgrade`
- Falco package operations now follow the official host-package approach and auto-detect `apt`, `dnf`, `yum`, or `zypper`

## Supported Hosts

- Linux hosts with `systemd`
- `x86_64` and `aarch64`
- Debian / Ubuntu via `apt`
- Amazon Linux / RHEL / Rocky / AlmaLinux / Fedora via `dnf` or `yum`
- openSUSE / SLES family via `zypper`

## Install

Copy both files to the target client machine, then run:

```bash
sudo bash install-falco-agent.sh \
  --server http://<gva-host>:8888 \
  --enroll-key <enroll-key> \
  --script-base-url http://<gva-host>:8888/falco/agent/install \
  --provider aws \
  --region ap-southeast-1
```

Optional Falco package behavior defaults:

```bash
sudo bash install-falco-agent.sh \
  --server http://<gva-host>:8888 \
  --enroll-key <enroll-key> \
  --falco-driver-choice modern_ebpf \
  --falcoctl-enabled no
```

## Service

The installer creates:

- `/etc/falco-agent/agent.env`
- `/opt/falco-agent/falco-agent.sh`
- `systemd` unit: `falco-agent.service`

The generated `agent.env` also stores default Falco package-install behavior:

- `FALCO_FRONTEND=noninteractive`
- `FALCO_DRIVER_CHOICE=<auto|kmod|ebpf|modern_ebpf|none>`
- `FALCOCTL_ENABLED=<yes|no>`

## Verify

```bash
sudo systemctl status falco-agent
sudo journalctl -u falco-agent -f
```

## Notes

- The Agent installer itself does not install Falco immediately; it only prepares the Agent.
- When the platform later sends `falco.install`, `falco.upgrade`, or `falco.rollback`, the runtime script will bootstrap the official Falco package repository for the detected package manager if needed.
