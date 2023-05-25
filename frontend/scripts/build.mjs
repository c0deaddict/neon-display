import esbuild from "esbuild";
// import liveServer from "live-server";
import fs from "fs/promises";

import indexHtmlPlugin from "./index-html-plugin.mjs";

let serve = false;

const entryPoint = "src/app.jsx";
const assetNames = "[name]-[hash]";

let options = {
  entryPoints: [entryPoint],
  outdir: "dist",
  bundle: true,
  format: "esm",
  target: "es2020",
  metafile: true,
  assetNames,
  entryNames: assetNames,
  plugins: [
    indexHtmlPlugin({
      entryPoint,
      htmlTemplate: await fs.readFile("src/index.html", "utf8"),
    }),
  ],
};

for (var arg of process.argv.slice(2)) {
  switch (arg) {
    case "--minify":
      options.minify = true;
      break;

    case "--watch":
      options.watch = {
        onRebuild(error, result) {
          if (error) console.error("watch build failed:", error);
        },
      };
      break;

    case "--serve":
      serve = true;
      break;

    default:
      console.error("unknown command line argument: ", arg);
      break;
  }
}

await esbuild.build(options);

// TODO: liveServer can't be build with npmlock2nix, find a different lib to do this.
// if (serve) {
//   liveServer.start({
//     port: 3000,
//     host: "127.0.0.1",
//     root: options.outdir,
//     open: false,
//   });
// }
