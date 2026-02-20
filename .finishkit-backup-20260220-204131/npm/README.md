# restless-uac

This is an npm wrapper for the **restless** Go CLI.

It downloads a matching prebuilt binary from GitHub Releases on install.

## Install
```sh
npm i -g restless-uac
restless --help
```

## Troubleshooting
- Ensure a GitHub Release exists for the same version as this package.
- Installer expects assets named like:
  `restless_<version>_<os>_<arch>.tar.gz`
