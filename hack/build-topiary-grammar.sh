#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
out="${1:-$repo_root/.topiary/tree-sitter-dang.so}"
tmpdir="$(mktemp -d)"

cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

mkdir -p "$(dirname "$out")"

cc -fPIC -I"$repo_root/treesitter/src" \
  -c "$repo_root/treesitter/src/parser.c" \
  -o "$tmpdir/parser.o"
cc -fPIC -I"$repo_root/treesitter/src" \
  -c "$repo_root/treesitter/src/scanner.c" \
  -o "$tmpdir/scanner.o"

case "$(uname -s)" in
  Darwin)
    cc -dynamiclib "$tmpdir/parser.o" "$tmpdir/scanner.o" -o "$out"
    ;;
  *)
    cc -shared "$tmpdir/parser.o" "$tmpdir/scanner.o" -o "$out"
    ;;
esac

printf 'wrote %s\n' "$out"
