#!/usr/bin/env bash
# Download all vendor libraries for offline use.
# Run from the project root or web-vue3/ directory:
#   bash web-vue3/scripts/download-vendor.sh
#
# To update a library, change the pinned version in this file and re-run.
# No npm or build tools required.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VENDOR="$SCRIPT_DIR/../vendor"

mkdir -p "$VENDOR/css" "$VENDOR/webfonts" "$VENDOR/js" "$VENDOR/js/bundles"

ESM="https://esm.sh"
UNPKG="https://unpkg.com"
CDNJS="https://cdnjs.cloudflare.com/ajax/libs"

# ── Pinned versions ────────────────────────────────────────────────────────
VUE_VER="3"
VUE_ROUTER_VER="4"
ELEMENT_PLUS_VER="2"
EP_ICONS_VER="2"
AXIOS_VER="1"
XTERM_VER="5"
XTERM_FIT_VER="0.10"
XTERM_WEBLINKS_VER="0.11"
CM_VER="6"
ZMODEM_VER="0.1.10"
FA_VER="6.5.0"
SFC_LOADER_VER="latest"

# ── Helpers ────────────────────────────────────────────────────────────────
fetch() {
  local url="$1" dest="$2"
  echo "  GET $url"
  curl -fsSL "$url" -o "$dest"
}

# Download a package from esm.sh with ?bundle, then fetch the actual .bundle.mjs
# files it references, rewrite /node/* → relative path, and write a clean local shim.
# NOTE: esm.sh minifies output so there is NO space between 'from' and the quote.
bundle_esm() {
  local pkg="$1" out_shim="$2" extra_flags="${3:-}"
  local tmp=$(mktemp)
  fetch "${ESM}/${pkg}?bundle${extra_flags:+&$extra_flags}" "$tmp"

  # Extract all .bundle.mjs paths referenced by the shim (quoted strings ending in bundle.mjs)
  local -a paths
  mapfile -t paths < <(grep -o '"[^"]*bundle\.mjs"' "$tmp" | tr -d '"' | sort -u | grep '^/')

  local shim_content=""
  local first_fname=""
  for path in "${paths[@]}"; do
    local fname; fname=$(basename "$path")
    [ -z "$first_fname" ] && first_fname="$fname"
    local bundle_url="${ESM}${path}"
    local bundle_dest="$VENDOR/js/bundles/$fname"
    echo "    → bundles/$fname"
    curl -fsSL "$bundle_url" -o "$bundle_dest"
    # Rewrite absolute /node/* references to relative paths (no space before quote in minified output)
    sed -i \
      's|from"/node/process\.mjs"|from"./process.mjs"|g;
       s|from"/node/events\.mjs"|from"./events.mjs"|g;
       s|from"/node/tty\.mjs"|from"./tty.mjs"|g;
       s|from"/node/async_hooks\.mjs"|from"./async_hooks.mjs"|g' "$bundle_dest"
    shim_content+="export * from \"./bundles/${fname}\";"$'\n'
  done
  # Re-export default if the shim declares one
  if grep -q 'export { default }' "$tmp" && [ -n "$first_fname" ]; then
    shim_content+="export { default } from \"./bundles/${first_fname}\";"$'\n'
  fi

  printf '%s' "$shim_content" > "$out_shim"
  rm -f "$tmp"
}

echo ""
echo "=== CSS ==="
fetch "$UNPKG/element-plus/dist/index.css"                               "$VENDOR/css/element-plus.css"
fetch "$UNPKG/@xterm/xterm/css/xterm.css"                                "$VENDOR/css/xterm.css"
fetch "$CDNJS/font-awesome/${FA_VER}/css/all.min.css"                    "$VENDOR/css/fontawesome.min.css"

echo ""
echo "=== Font Awesome webfonts ==="
for font in fa-solid-900 fa-regular-400 fa-brands-400; do
  fetch "$CDNJS/font-awesome/${FA_VER}/webfonts/${font}.woff2" "$VENDOR/webfonts/${font}.woff2"
  fetch "$CDNJS/font-awesome/${FA_VER}/webfonts/${font}.ttf"   "$VENDOR/webfonts/${font}.ttf"
done

echo ""
echo "=== Simple self-contained ESM files ==="
fetch "$UNPKG/vue@${VUE_VER}/dist/vue.esm-browser.prod.js"               "$VENDOR/js/vue.esm.js"
fetch "$UNPKG/vue-router@${VUE_ROUTER_VER}/dist/vue-router.esm-browser.prod.js" "$VENDOR/js/vue-router.esm.js"
fetch "$UNPKG/vue3-sfc-loader/dist/vue3-sfc-loader.esm.js"               "$VENDOR/js/vue3-sfc-loader.esm.js"

echo ""
echo "=== element-plus (full single-file bundle, external=vue) ==="
# index.full.mjs is element-plus's own pre-built single-file bundle
fetch "$UNPKG/element-plus/dist/index.full.mjs"                           "$VENDOR/js/element-plus.esm.js"

echo ""
echo "=== Node polyfills (needed by xterm, codemirror lang packs) ==="
fetch "${ESM}/node/async_hooks.mjs"                                       "$VENDOR/js/bundles/async_hooks.mjs"
fetch "${ESM}/node/events.mjs"                                            "$VENDOR/js/bundles/events.mjs"
fetch "${ESM}/node/tty.mjs"                                               "$VENDOR/js/bundles/tty.mjs"
fetch "${ESM}/node/process.mjs"                                           "$VENDOR/js/bundles/process.mjs"
# Rewrite absolute /node/ references to relative paths (esm.sh minifies: no space before quote)
for f in "$VENDOR/js/bundles/process.mjs" "$VENDOR/js/bundles/events.mjs"; do
  sed -i \
    's|from"/node/async_hooks\.mjs"|from"./async_hooks.mjs"|g;
     s|from"/node/events\.mjs"|from"./events.mjs"|g;
     s|from"/node/tty\.mjs"|from"./tty.mjs"|g;
     s|from"/node/process\.mjs"|from"./process.mjs"|g' "$f"
done

echo ""
echo "=== esm.sh bundles ==="
bundle_esm "@element-plus/icons-vue@${EP_ICONS_VER}" "$VENDOR/js/element-plus-icons.esm.js" "external=vue"
bundle_esm "@xterm/xterm@${XTERM_VER}"                "$VENDOR/js/xterm.esm.js"
bundle_esm "@xterm/addon-fit@${XTERM_FIT_VER}"        "$VENDOR/js/xterm-addon-fit.esm.js"    "external=@xterm/xterm"
bundle_esm "@xterm/addon-web-links@${XTERM_WEBLINKS_VER}" "$VENDOR/js/xterm-addon-web-links.esm.js" "external=@xterm/xterm"
bundle_esm "zmodem.js@${ZMODEM_VER}"                  "$VENDOR/js/zmodem.esm.js"

echo ""
echo "=== CodeMirror 6 core sub-packages (each separately, cross-externalized) ==="
CM_BASE="external=@codemirror/state"
CM_VIEW="external=@codemirror/state,@codemirror/view"
CM_LANG="external=@codemirror/state,@codemirror/view,@codemirror/language"
CM_FULL="external=@codemirror/state,@codemirror/view,@codemirror/language,@codemirror/commands"
CM_ALL="external=@codemirror/state,@codemirror/view,@codemirror/language,@codemirror/commands,@codemirror/autocomplete,@codemirror/lint,@codemirror/search"
bundle_esm "@codemirror/state@${CM_VER}"        "$VENDOR/js/codemirror-state.esm.js"
bundle_esm "@codemirror/view@${CM_VER}"         "$VENDOR/js/codemirror-view.esm.js"         "$CM_BASE"
bundle_esm "@codemirror/language@${CM_VER}"     "$VENDOR/js/codemirror-language.esm.js"     "$CM_VIEW"
bundle_esm "@codemirror/commands@${CM_VER}"     "$VENDOR/js/codemirror-commands.esm.js"     "$CM_LANG"
bundle_esm "@codemirror/autocomplete@${CM_VER}" "$VENDOR/js/codemirror-autocomplete.esm.js" "$CM_FULL"
bundle_esm "@codemirror/lint@${CM_VER}"         "$VENDOR/js/codemirror-lint.esm.js"         "$CM_LANG"
bundle_esm "@codemirror/search@${CM_VER}"       "$VENDOR/js/codemirror-search.esm.js"       "$CM_LANG"
bundle_esm "codemirror@${CM_VER}"               "$VENDOR/js/codemirror.esm.js"              "$CM_ALL"

echo ""
echo "=== CodeMirror language packs (external: all @codemirror/* core) ==="
CM_EXT="external=@codemirror/state,@codemirror/view,@codemirror/language,@codemirror/commands,@codemirror/autocomplete,@codemirror/lint,@codemirror/search"
bundle_esm "@codemirror/lang-javascript@${CM_VER}" "$VENDOR/js/codemirror-lang-javascript.esm.js" "$CM_EXT"
bundle_esm "@codemirror/lang-python@${CM_VER}"     "$VENDOR/js/codemirror-lang-python.esm.js"     "$CM_EXT"
bundle_esm "@codemirror/lang-html@${CM_VER}"       "$VENDOR/js/codemirror-lang-html.esm.js"       "$CM_EXT"
bundle_esm "@codemirror/lang-css@${CM_VER}"        "$VENDOR/js/codemirror-lang-css.esm.js"        "$CM_EXT"
bundle_esm "@codemirror/lang-json@${CM_VER}"       "$VENDOR/js/codemirror-lang-json.esm.js"       "$CM_EXT"
bundle_esm "@codemirror/lang-markdown@${CM_VER}"   "$VENDOR/js/codemirror-lang-markdown.esm.js"   "$CM_EXT"

echo ""
echo "=== Done! ==="
du -sh "$VENDOR"
