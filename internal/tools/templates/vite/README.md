# {{.AppName}}

This Vite app was created with Layered Code.

## Structure

- **src/** - Source code for your application
- **build/** - Compiled/built files for deployment (gitignored)
- **vite.config.js** - Vite configuration
- **package.json** - Dependencies and scripts
- **.layered-code/** - Layered Code metadata (gitignored)
- **.layered.json** - Layered Code configuration

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the development server:
   ```bash
   npm run dev
   ```

3. Build for production:
   ```bash
   npm run build
   ```

## Development

- Edit `src/main.js` - Main application entry point
- Edit `src/style.css` - Application styles
- Edit `src/index.html` - HTML template

Vite provides Hot Module Replacement (HMR) for instant updates during development.

## Deployment

### Netlify
- Build command: `npm run build`
- Publish directory: `build`

### Other platforms
- Build your app using `npm run build`
- Deploy the contents of the `build/` directory

## Notes

The `.layered-code/` directory and `build/` directory are gitignored by default to keep your repository clean and prevent accidental deployment of development files.