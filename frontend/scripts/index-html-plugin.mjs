import path from "path";
import fs from "fs/promises";

export default function indexHtmlPlugin({ htmlTemplate, entryPoint }) {
  const entryPointRegex = new RegExp(
    `^${path.parse(entryPoint).name}(-[A-Z0-9]{8})?.js$`
  );
  return {
    name: "index-html",
    setup(build) {
      build.onEnd(async (result) => {
        let outputs = [];
        if (result.outputFiles) {
          outputs = result.outputFiles.map((out) => out.path);
        } else {
          outputs = Object.keys(result.metafile.outputs).map((p) =>
            path.resolve(process.cwd(), p)
          );
        }

        let stylesheets = [];
        let entryPoint = null;
        for (let file of outputs) {
          let relpath = path.relative(build.initialOptions.outdir, file);
          if (relpath.endsWith(".css")) {
            stylesheets.push(`<link rel="stylesheet" href="${relpath}">`);
          } else if (relpath.match(entryPointRegex)) {
            if (entryPoint !== null) {
              console.warn("Multiple outputs match the entrypoint regex!");
            }
            entryPoint = `<script type="module" src="${relpath}" defer=""></script>`;
          }
        }

        stylesheets = stylesheets.join("\n");
        let replacements = { stylesheets, entryPoint };

        let contents = htmlTemplate.replace(
          /<!-- @(\w+)@ -->/g,
          (_, p1) => replacements[p1]
        );
        const outpath = path.join(build.initialOptions.outdir, "index.html");
        if (result.outputFiles) {
          result.outputFiles.push({
            path: outpath,
            contents: new TextEncoder().encode(contents),
            text: contents,
          });
        } else {
          await fs.writeFile(outpath, contents, "utf8");
        }
      });
    },
  };
}
