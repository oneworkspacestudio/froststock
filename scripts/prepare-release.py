"""Create the managed update manifest and immutable checksum list."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
REPOSITORY = "oneworkspacestudio/froststock"
MINIMUM_MANAGER = "0.3.0-preview.140"
ASSETS = {
    "froststock": "Froststock.exe",
    "froststock-contract": "compatibility.json",
    "froststock-guide": "README.md",
    "froststock-license": "LICENSE",
    "froststock-notices": "NOTICE",
}


def digest(file):
    checksum = hashlib.sha256()
    with file.open("rb") as stream:
        while block := stream.read(4 << 20):
            checksum.update(block)
    return checksum.hexdigest()


def write_json(file, value):
    file.write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")


def prepare(output, commit, sdk):
    if not re.fullmatch(r"[a-f0-9]{40}", commit):
        raise ValueError("invalid source commit")
    executable = output / "Froststock.exe"
    version = subprocess.check_output([str(executable), "--version"], text=True).strip()
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?", version):
        raise ValueError("invalid Froststock version")
    tag = "v" + version
    contract = json.loads((ROOT / "workspace-app.json").read_bytes())
    if set(contract["packages"]) != set(ASSETS):
        raise ValueError("unexpected Froststock package namespace")
    app = dict(contract["app"], version=version,
               source=f"https://github.com/{REPOSITORY}/releases/tag/{tag}")
    write_json(output / "compatibility.json", {
        "schema": 1, "id": "froststock", "version": version,
        "minimumWorkspaceVersion": MINIMUM_MANAGER, "workspaceInstallerIntegration": 22,
        "route": app["route"], "oidc": app["oidc"], "executable": app["executable"],
        "health": app["health"], "storage": "my-drive-froststock-v1",
    })
    packages = {}
    for key, filename in ASSETS.items():
        asset = output / filename
        packages[key] = {"asset": filename, "file": contract["packages"][key],
                         "sha256": digest(asset), "size": asset.stat().st_size}
    write_json(output / "workspace-update.json", {
        "schema": 2, "integration": 3, "repository": REPOSITORY, "tag": tag,
        "minimumManager": MINIMUM_MANAGER, "apps": [app], "packages": packages,
    })
    write_json(output / "build-info.json", {
        "schema": 1, "repository": REPOSITORY, "tag": tag, "sourceCommit": commit,
        "workspaceAppHost": sdk, "platform": "windows-amd64",
    })
    assets = sorted(output.iterdir())
    with (output / "SHA256SUMS.txt").open("w", encoding="utf-8", newline="\n") as checksums:
        for asset in assets:
            if not asset.is_file() or asset.name == "SHA256SUMS.txt":
                raise ValueError("release output must be a fresh directory containing only assets")
            checksums.write(f"{digest(asset)}  {asset.name}\n")
    print(tag)


if __name__ == "__main__":
    prepare(Path(sys.argv[1]).resolve(), sys.argv[2], sys.argv[3])
