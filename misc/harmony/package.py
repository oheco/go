#!/usr/bin/env python3
"""Package a signed native GOROOT, preserving standard names and permissions."""

import hashlib
from pathlib import Path
import subprocess
import sys
import tarfile
import tempfile


def main():
    root, output = (Path(arg).resolve() for arg in sys.argv[1:])
    output.mkdir(parents=True, exist_ok=True)
    version = (root / "VERSION").read_text().splitlines()[0]
    archive = output / f"{version}.ohos-arm64.tar.gz"
    checksum = archive.with_suffix(archive.suffix + ".sha256")
    if archive.exists() or checksum.exists():
        sys.exit(f"Output already exists: {archive} or {checksum}")

    for name in ("bin/go", "bin/gofmt", "pkg/tool/ohos_arm64/compile",
                 "pkg/tool/ohos_arm64/link", "pkg/tool/ohos_arm64/cgo"):
        if not (root / name).is_file():
            sys.exit(f"Missing native tool: {name}")

    # The host share presents every file as mode 0777. Recover source modes
    # from Git; otherwise archives would make all Go source executable.
    staged = subprocess.check_output(["git", "-C", str(root), "ls-files", "-s", "-z"])
    executable = set()
    for entry in staged.decode().split("\0"):
        if entry.startswith("100755 "):
            executable.add(entry.split("\t", 1)[1])

    def metadata(info):
        name = info.name.removeprefix("go/")
        if "__pycache__" in Path(name).parts:
            return None
        if name.startswith("misc/harmony/testdata/"):
            if name.endswith((".test", "/libgoharmony.so", "/libgoharmony.h")):
                return None
        info.uid = info.gid = 0
        info.uname = info.gname = "root"
        if info.isdir() or info.issym():
            info.mode = 0o755 if info.isdir() else 0o777
        elif info.isfile():
            is_tool = name.startswith(("bin/", "pkg/tool/ohos_arm64/"))
            is_script = name == "src/make.sh" or (
                name.startswith("misc/harmony/") and name.endswith((".sh", ".py")))
            info.mode = 0o755 if is_tool or is_script or name in executable else 0o644
        else:
            raise ValueError(f"Unexpected special file in GOROOT: {name}")
        return info

    members = (
        "VERSION LICENSE PATENTS README.md SECURITY.md CONTRIBUTING.md go.env "
        "api bin doc lib misc/harmony pkg/include pkg/tool/ohos_arm64 src test"
    ).split()
    # A failed read/compression never leaves an apparently complete archive.
    with tempfile.TemporaryDirectory(prefix=".go-package-", dir=output) as work:
        staged_archive = Path(work) / archive.name
        with tarfile.open(staged_archive, "w:gz", format=tarfile.GNU_FORMAT) as tar:
            for name in members:
                tar.add(root / name, arcname="go/" + name, filter=metadata)
        digest = hashlib.sha256()
        with staged_archive.open("rb") as file:
            for block in iter(lambda: file.read(1024 * 1024), b""):
                digest.update(block)
        # Exclusive creation also protects against a concurrent packaging run.
        with archive.open("xb") as dest, staged_archive.open("rb") as source:
            try:
                for block in iter(lambda: source.read(1024 * 1024), b""):
                    dest.write(block)
            except BaseException:
                archive.unlink()
                raise
        with checksum.open("x") as file:
            file.write(f"{digest.hexdigest()}  {archive.name}\n")
    print(f"Packaged {archive}")


if __name__ == "__main__":
    main()
