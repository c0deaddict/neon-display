When installing or updating packages, make sure to include the optional dependencies for esbuild:

```bash
npm i esbuild --include=optional
```

Otherwise the arm64 binaries are missing in the package-lock.json
