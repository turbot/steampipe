# Steampipe Dashboards

This will allow you to develop the dashboards UI locally. To get started:

- Ensure you have `steampipe` installed - https://steampipe.io/downloads
- Ensure you have `node` LTS installed - https://nodejs.org/en/download/
- Ensure you have `yarn` installed - https://yarnpkg.com/getting-started/install

## Start Steampipe

- Start your local Steampipe service, and dashboard service, with `steampipe dashboard` (ensure you have navigated to your mod folder with `cd <mod_folder>` first)

## UI Setup

- Navigate to the dashboard UI with `cd ui/dashboard`
- Run `yarn install`
- Run `yarn start`
- The browser will open at http://localhost:3000
- Any changes you make to the source code will be reflected in the dashboard UI in realtime  

## Component library

To browse the dashboard component library do the following:

- `cd ui/dashboard`
- `yarn install`
- `yarn storybook`

Your default browser will open up the component library.

## Adding new `chart` types

- Navigate to `ui/dashboard/src/components/dashboards/charts`
- Add a folder for the new chart type and add `index.tsx`
- Ensure you call `registerComponent` before the export
- Import this new chart in `ui/dashboard/src/utils/registerComponents.ts`
- Add Storybook stories for your component in `index.stories.js`

- ## Adding new `flow` types

- Navigate to `ui/dashboard/src/components/dashboards/flows`
- Add a folder for the new flow type and add `index.tsx`
- Ensure you call `registerComponent` before the export
- Import this new flow in `ui/dashboard/src/utils/registerComponents.ts`
- Add Storybook stories for your component in `index.stories.js`

- ## Adding new `hierarchy` types

- Navigate to `ui/dashboard/src/components/dashboards/hierarchies`
- Add a folder for the new hierarchy type and add `index.tsx`
- Ensure you call `registerComponent` before the export
- Import this new hierarchy in `ui/dashboard/src/utils/registerComponents.ts`
- Add Storybook stories for your component in `index.stories.js`
