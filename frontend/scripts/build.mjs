import esbuild from 'esbuild';
import htmlPlugin from '@chialab/esbuild-plugin-html';
import http from 'http';

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
        else console.log('watch build succeeded:', result);
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

if (!serve) {
  await esbuild.build(options);
} else {
  // https://esbuild.github.io/api/#customizing-server-behavior
  const {host, port} = await esbuild.serve({
    servedir: options.outdir
  }, options);

  http.createServer((req, res) => {
    const options = {
      hostname: host,
      port: port,
      path: req.url,
      method: req.method,
      headers: req.headers,
    };

    const proxyReq = http.request(options, proxyRes => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res, { end: true });
    });

    req.pipe(proxyReq, { end: true });
  }).listen(3000);

  console.log('esbuild serve running on http://localhost:3000');
}
