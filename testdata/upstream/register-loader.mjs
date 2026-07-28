// Registers loader.mjs as a module customization hook.
// Usage: node --import ./register-loader.mjs runner.mjs --stage2 <files>
import { register } from 'node:module';
register('./loader.mjs', import.meta.url);
