// Run from the Plenti project: node scripts/update-pattr.mjs
import { copyFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const pattrRoot = resolve(projectRoot, '../pattr');
const npm = process.platform === 'win32' ? 'npm.cmd' : 'npm';

console.log(`Building Pattr in ${pattrRoot}...`);
execFileSync(npm, ['run', 'build'], { cwd: pattrRoot, stdio: 'inherit' });

const source = resolve(pattrRoot, 'pattr.min.js');
const destination = resolve(projectRoot, 'defaults/core/pattr.js');
copyFileSync(source, destination);
console.log(`Updated ${destination} from ${source}.`);
