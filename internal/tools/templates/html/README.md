# {{.AppName}}

This app was created with Layered Code.

## Structure

- **src/** - Source code for your application
- **build/** - Compiled/built files for deployment (gitignored)
- **package.json** - Build scripts and project metadata
- **.layered-code/** - Layered Code metadata (gitignored)
- **.layered.json** - Layered Code configuration

## Getting Started

1. Add your source code to the `src/` directory
2. Run `npm run build` to copy files to the `build/` directory
3. Deploy to your hosting platform (see below)

## Deployment

### Netlify
- Build command: `npm run build`
- Publish directory: `build`

### Other platforms
- Build your app using `npm run build`
- Deploy the contents of the `build/` directory

## Notes

The `.layered-code/` directory and `build/` directory are gitignored by default to keep your repository clean and prevent accidental deployment of development files.