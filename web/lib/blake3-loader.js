// BLAKE3 loader - wraps the ES module and exposes it globally
import init, { create_hasher } from './blake3_js.js';

// Initialize the WASM module
await init('./lib/blake3_js_bg.wasm');

// Create a global blake3 object compatible with the existing API
window.blake3 = {
    createHash: function() {
        const hasher = create_hasher();

        return {
            update: function(data) {
                hasher.update(data);
            },
            digest: function() {
                // BLAKE3 outputs 32 bytes by default
                const out = new Uint8Array(32);
                hasher.digest(out);
                return out;
            }
        };
    }
};
