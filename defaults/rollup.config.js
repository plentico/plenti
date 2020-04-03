import svelte from 'rollup-plugin-svelte';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import livereload from 'rollup-plugin-livereload';
import { terser } from 'rollup-plugin-terser';

const production = !process.env.ROLLUP_WATCH;

export default [{
    input: 'layout/ejected/main.js',
    output: {
        sourcemap: true,
        format: 'esm',
        name: 'app',
        dir: 'public/build/spa'
    },
    plugins: [
        svelte({
            dev: !production,
            css: css => {
                css.write('public/build/bundle.css');
            }
        }),
        resolve({
            browser: true,
            dedupe: importee => importee === 'svelte' || importee.startsWith('svelte/')
        }),
        commonjs(),
        !production && serve(),
        !production && livereload('public'),
        production && terser()
    ],
    watch: {
        clearScreen: false
    }
},
{
    input: 'layout/ejected/main.js',
    output: {
        sourcemap: true,
        format: 'cjs',
        name: 'app',
        dir: 'public/build/static'
    },
    plugins: [
        svelte({
            generate: 'ssr'
        }),
    ]
}];

function serve() {
    let started = false;

    return {
        writeBundle() {
            if (!started) {
                started = true;

                require('child_process').spawn('npm', ['run', 'start', '--', '--dev'], {
                    stdio: ['ignore', 'inherit', 'inherit'],
                    shell: true
                });
            }
        }
    };
}