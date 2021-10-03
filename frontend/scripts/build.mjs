import esbuild from 'esbuild';
import htmlPlugin from '@chialab/esbuild-plugin-html';
import liveServer from 'live-server';

let serve = false;

let options = {
  entryPoints: ['src/index.html'],
  outdir: 'dist',
  bundle: true,
  plugins: [
    htmlPlugin({
      scriptsTarget: 'es2015',
      modulesTarget: 'es2020',
    }),
  ],
};

for (var arg of process.argv.slice(2)) {
  switch (arg) {
  case '--minify':
    options.minify = true;
    break;

  case '--watch':
    options.watch = {
      onRebuild(error, result) {
        if (error) console.error('watch build failed:', error);
      },
    };
    break;

  case '--serve':
    serve = true;
    break;

  default:
    console.error('unknown command line argument: ', arg);
    break;
  }
}

await esbuild.build(options);

if (serve) {
  liveServer.start({
    port: 3000,
    host: "127.0.0.1",
    root: options.outdir,
    open: false,
  });
}
