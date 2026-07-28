// Node module customization hook: redirects the CDN URLs that the vendored
// pos-tagger.js imports to the locally installed, pinned wink-pos-tagger.
// The src entry is CommonJS; the ESM import path in pos-tagger.js picks it up
// as module.default (the tagger factory). The browser dist bundle is not used:
// it needs window/document. Used only by the stage-2 delta measurement; see
// README.md.
const local = new URL('./node_modules/wink-pos-tagger/src/wink-pos-tagger.js', import.meta.url).href;

export async function resolve(specifier, context, next) {
  if (specifier.startsWith('https://')) {
    return { url: local, shortCircuit: true };
  }
  return next(specifier, context);
}
