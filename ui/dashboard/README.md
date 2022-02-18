# Steampipe Reports

This will allow you test out reports locally. To get started:

- Ensure you have `go` installed - https://golang.org/doc/install
- Ensure you have `node` LTS installed - https://nodejs.org/en/download/
- Ensure you have `yarn` installed - https://yarnpkg.com/getting-started/install
- Start your local Steampipe service with `steampipe service start`
- Run `build.sh` (you only need to do this if the ui has been updated)
- Run `cd server`
- Run `make`
- Open your browser at http://localhost:5000
- Edit `server/report.yml` to update the report
- Enjoy!

## Component library

To browse the reporting component library do the following:

- `cd ui`
- `yarn install`
- `yarn storybook`

Your default browser will open up the component library.

## TODO & Questions

* Panel idea - Table of contents
* Panel idea - Regional world map
* How to display false in table data cells (currently empty string)?
* Empty tables need a filler / default row (e.g. None).
* Zero is often a special case - e.g. no public buckets. How do we handle that?
* Use column types to format data tables (e.g. integers should align right by default)
* How do we jump to / link to / hide extra data? e.g. Buckets by region, I want to jump to the one bad bucket ... how can I do that?
* Can I get to the SQL for a panel?
* Tables should show partial data, with a click to see more / expand option? (e.g. very long bucket list)
* Component loading placeholders should match their shape (e.g. line chart, pie chart, etc)

## Outline Design

- `steampipe service start` will additionally start a reporting server
- The reporting server consists of a go server (with REST API) and a React Single Page Application (SPA) for local report development/testing
- Assume the reporting server SPA is accessible via https://localhost:5000 `TODO protocol (http vs https) & port TBD`
- Navigating to the reporting server SPA will allow the user to choose a workspace on their machine
  `TODO are these discoverable via Steampipe? We possibly need a "default"/"scratch" workspace path that we can load?`
- For the selected workspace, provide a list of available reports
  `TODO how do we list reports for a workspace? steampipe report list etc?`
- Default the report selection to the "default"/"main" report
- The server then needs to run the report, which largely consists of 3 pieces:
    - `Parsing` - `steampipe report load?` We need to get back definition of the report over a go channel. This will allow the
      SPA to build the skeleton of the report. Time To First Pixel (TTFP) is key here - even if we don't have the required data
      to render all of the report, dependent on the report target we can lay out the pieces with placeholders etc. Parsing the
      report will also provide the query data requirements that can be passed to execution
    - `Query Execution` - We require a way to execute the queries identified during the parsing phase. These queries will be passed to
      an execution engine. Data will be streamed out via Postgres over go channels. Available data will be streamed to the render engine as it becomes avaiable.
    - `Rendering` - The initial reporting targets are: `web`, `slack` and `email`. We may also require `terminal` as a target (for ASCII reporting)
      For `slack` and `email`, all data is required to render the report. `web` can provide a gradual experience, to achieve minimum TTFP.
      The initial report skeleton can be rendered as soon as its available, including markdown. Placeholders can be rendered in place of panels
      whose data is pending. As data is steamed out of the execution engine, it can be passed to the render engine and the report can fill out progressively.
- `TODO if the user edits report HCL, how do we effeciently re-run this process? Talk more in depth with Kai about diffing etc.
  The desired output of a change to the HCL is any updated panels and if relevant, any queries required to stream data out of the execution engine`