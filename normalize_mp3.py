#!/usr/bin/env python3
"""
normalize_mp3.py — Normalize the loudness of MP3 files to a target level.
Works with Python 3.13+. Uses ffmpeg directly (no pydub).

Usage:
    python normalize_mp3.py [OPTIONS] FILE_OR_DIR [FILE_OR_DIR ...]

Examples:
    # Normalize a single file (overwrites original)
    python normalize_mp3.py song.mp3

    # Normalize all MP3s in a folder, saving to an output folder
    python normalize_mp3.py ./music/ --output ./normalized/

    # Normalize multiple files to -16 LUFS (podcast standard)
    python normalize_mp3.py ./music/ --target -16

    # Dry run — show what would change without writing files
    python normalize_mp3.py ./music/ --dry-run

Dependencies:
    pip install pyloudnorm numpy
    Also requires ffmpeg: https://ffmpeg.org/download.html
"""

import argparse
import sys
import shutil
import subprocess
import tempfile
import os
from pathlib import Path


def check_dependencies():
    missing = []
    try:
        import pyloudnorm
    except ImportError:
        missing.append("pyloudnorm")
    try:
        import numpy
    except ImportError:
        missing.append("numpy")

    if missing:
        print(f"Missing dependencies: {', '.join(missing)}")
        print(f"Install with: pip install {' '.join(missing)}")
        sys.exit(1)

    if not shutil.which("ffmpeg"):
        print("ffmpeg not found. Please install it: https://ffmpeg.org/download.html")
        print("  macOS:   brew install ffmpeg")
        print("  Ubuntu:  sudo apt install ffmpeg")
        print("  Windows: https://ffmpeg.org/download.html")
        sys.exit(1)


def decode_to_pcm(input_path: Path):
    """Use ffmpeg to decode MP3 to raw 32-bit float PCM. Returns (samples, rate, channels)."""
    import numpy as np

    # First, probe for sample rate and channels
    probe = subprocess.run(
        ["ffmpeg", "-i", str(input_path), "-hide_banner"],
        capture_output=True, text=True
    )
    # Parse rate and channels from ffmpeg stderr output
    rate, channels = 44100, 2  # defaults
    for line in probe.stderr.splitlines():
        if "Audio:" in line:
            parts = line.split(",")
            for part in parts:
                part = part.strip()
                if "Hz" in part:
                    try:
                        rate = int(part.replace("Hz", "").strip())
                    except ValueError:
                        pass
                if part in ("mono", "stereo"):
                    channels = 1 if part == "mono" else 2
                elif "channels" in part:
                    try:
                        channels = int(part.split()[0])
                    except ValueError:
                        pass

    # Decode to raw PCM float32
    result = subprocess.run(
        ["ffmpeg", "-i", str(input_path),
         "-f", "f32le", "-acodec", "pcm_f32le",
         "-ar", str(rate), "-ac", str(channels),
         "pipe:1", "-hide_banner", "-loglevel", "error"],
        capture_output=True
    )
    if result.returncode != 0:
        raise RuntimeError(result.stderr.decode())

    samples = np.frombuffer(result.stdout, dtype=np.float32)
    if channels > 1:
        samples = samples.reshape((-1, channels))
    return samples, rate, channels


def encode_from_pcm(samples, rate: int, channels: int, output_path: Path, bitrate: str = "320k"):
    """Use ffmpeg to encode raw PCM float32 back to MP3."""
    import numpy as np

    output_path.parent.mkdir(parents=True, exist_ok=True)
    raw = samples.astype(np.float32).tobytes()

    with tempfile.NamedTemporaryFile(suffix=".mp3", delete=False) as tmp:
        tmp_path = tmp.name

    try:
        result = subprocess.run(
            ["ffmpeg", "-y",
             "-f", "f32le", "-ar", str(rate), "-ac", str(channels),
             "-i", "pipe:0",
             "-acodec", "libmp3lame", "-b:a", bitrate,
             tmp_path, "-hide_banner", "-loglevel", "error"],
            input=raw, capture_output=True
        )
        if result.returncode != 0:
            raise RuntimeError(result.stderr.decode())
        shutil.move(tmp_path, str(output_path))
    except Exception:
        if os.path.exists(tmp_path):
            os.unlink(tmp_path)
        raise


def normalize_file(input_path: Path, output_path: Path, target_lufs: float, dry_run: bool):
    import pyloudnorm as pyln

    try:
        samples, rate, channels = decode_to_pcm(input_path)
    except Exception as e:
        print(f"  ERROR reading {input_path.name}: {e}")
        return False

    meter = pyln.Meter(rate)
    current_lufs = meter.integrated_loudness(samples)

    if current_lufs == float("-inf"):
        print(f"  SKIP {input_path.name} — silence detected")
        return False

    gain_db = target_lufs - current_lufs
    gain_linear = 10 ** (gain_db / 20.0)
    print(f"  {input_path.name}: {current_lufs:.1f} LUFS → {target_lufs:.1f} LUFS  (gain: {gain_db:+.1f} dB)")

    if dry_run:
        return True

    normalized = samples * gain_linear

    # Clip to prevent distortion
    import numpy as np
    normalized = np.clip(normalized, -1.0, 1.0)

    try:
        encode_from_pcm(normalized, rate, channels, output_path)
    except Exception as e:
        print(f"  ERROR writing {output_path.name}: {e}")
        return False

    return True


def collect_mp3s(paths):
    result = []
    for p in paths:
        p = Path(p)
        if p.is_dir():
            result.extend(sorted(p.rglob("*.mp3")))
        elif p.is_file() and p.suffix.lower() == ".mp3":
            result.append(p)
        else:
            print(f"Warning: skipping '{p}' (not an MP3 or directory)")
    return result


def main():
    parser = argparse.ArgumentParser(
        description="Normalize MP3 loudness to a target LUFS (Loudness Units relative to Full Scale).",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument("inputs", nargs="+", help="MP3 files or directories to process")
    parser.add_argument(
        "--target", type=float, default=-14.0,
        help="Target loudness in LUFS (default: -14, streaming standard). "
             "Use -16 for podcasts, -23 for broadcast.",
    )
    parser.add_argument(
        "--output", "-o", type=str, default=None,
        help="Output directory. If omitted, files are overwritten in place.",
    )
    parser.add_argument(
        "--dry-run", action="store_true",
        help="Print what would happen without writing any files.",
    )
    args = parser.parse_args()

    check_dependencies()

    mp3_files = collect_mp3s(args.inputs)
    if not mp3_files:
        print("No MP3 files found.")
        sys.exit(1)

    out_dir = Path(args.output) if args.output else None
    mode = "DRY RUN — " if args.dry_run else ""
    print(f"\n{mode}Normalizing {len(mp3_files)} file(s) to {args.target} LUFS\n")

    success = 0
    for src in mp3_files:
        dest = (out_dir / src.name) if out_dir else src
        if normalize_file(src, dest, args.target, args.dry_run):
            success += 1

    print(f"\nDone — {success}/{len(mp3_files)} file(s) processed.")


if __name__ == "__main__":
    main()
    