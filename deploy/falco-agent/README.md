# Falco Agent Installer

This directory contains a minimal first-phase installer for host-side Falco Agent.

## Files

- `install-falco-agent.sh`: installs the platform-side Falco Agent service on the client host
- `falco-agent.sh`: the runtime process of that Agent; it is not a separate Falco installer

## What It Does

- `install-falco-agent.sh` installs `falco-agent.service`, registers the host to the platform, and writes local config
- `falco-agent.sh` keeps heartbeats, reports status, pulls tasks, and reports task results
- Actual Falco package install/upgrade is triggered later by platform tasks such as `falco.install` and `falco.upgrade`

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

## Service

The installer creates:

- `/etc/falco-agent/agent.env`
- `/opt/falco-agent/falco-agent.sh`
- `systemd` unit: `falco-agent.service`

## Verify

```bash
sudo systemctl status falco-agent
sudo journalctl -u falco-agent -f
```
