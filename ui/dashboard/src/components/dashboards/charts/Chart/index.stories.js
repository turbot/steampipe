import Chart from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Charts/Chart",
  component: Chart,
  excludeStories: ["SingleTimeSeriesDefaults", "MultiTimeSeriesDefaults", "MultiTimeSeriesGroupedDefaults", "MultiTimeSeriesCrosstabDefaults"]
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} panelType="chart" />
);

export const DefaultsToColumn = Template.bind({});
DefaultsToColumn.args = {
  data: {
    columns: [
      { name: "Type", data_type: "TEXT" },
      { name: "Count", data_type: "INT8" },
    ],
    rows: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
};

export const SingleTimeSeriesDefaults = {
  data: {
    columns: [
      { name: "time", data_type: "TIMESTAMP" },
      { name: "count", data_type: "INT8" },
    ],
    rows: [
      { time: "2023-01-01T00:00:00.000", count: 20 },
      { time: "2023-02-01T00:00:00.000", count: 32 },
      { time: "2023-04-01T00:00:00.000", count: -15 },
      { time: "2023-05-01T00:00:00.000", count: 18 },
      { time: "2023-06-01T00:00:00.000", count: -9 },
      { time: "2023-12-01T00:00:00.000", count: 3 },
    ],
  },
};

const MultiTimeSeriesDataSample = {
  columns: [
    { name: "time", data_type: "TIMESTAMPTZ" },
    { name: "Income", data_type: "INT8" },
    { name: "Spending", data_type: "INT8" },
  ],
  rows: [
    { time: "2023-01-01T00:00:00.000Z", Income: 20, Spending: 0 },
    { time: "2023-02-01T00:00:00.000Z", Income: 18, Spending: 32 },
    { time: "2023-04-01T00:00:00.000Z", Income: 15, Spending: 3 },
    { time: "2023-05-01T00:00:00.000Z", Income: 18, Spending: 15 },
    { time: "2023-06-01T00:00:00.000Z", Income: 0, Spending: 9 },
    { time: "2023-09-01T00:00:00.000Z", Income: 7, Spending: 26 },
    { time: "2023-12-01T00:00:00.000Z", Income: 8, Spending: 3 },
  ],
};
const MultiTimeSeriesCrosstabDataSample = {
  columns: [
    { name: "time", data_type: "DATE" },
    { name: "series_name", data_type: "TEXT" },
    { name: "value", data_type: "INT8" },
  ],
  rows: [
    { time: "2023-01-01", series_name: "Income", value: 20},
    { time: "2023-01-01", series_name: "Spending", value: 0},
    { time: "2023-02-01", series_name: "Income", value: 18},
    { time: "2023-02-01", series_name: "Spending", value: 32},
    { time: "2023-04-01", series_name: "Income", value: 15},
    { time: "2023-04-01", series_name: "Spending", value: 3},
    { time: "2023-05-01", series_name: "Income", value: 18},
    { time: "2023-05-01", series_name: "Spending", value: 15},
    { time: "2023-06-01", series_name: "Income", value: 0},
    { time: "2023-06-01", series_name: "Spending", value: 9},
    { time: "2023-09-01", series_name: "Income", value: 7},
    { time: "2023-09-01", series_name: "Spending", value: 26},
    { time: "2023-12-01", series_name: "Income", value: 8},
    { time: "2023-12-01", series_name: "Spending", value: 3},
  ],
};

export const MultiTimeSeriesDefaults = {
  data: MultiTimeSeriesDataSample,
  properties: {
    series: {
      Income: {
        color: "green"
      },
      Spending: {
        color: "red"
      },
    }
  },
};

export const MultiTimeSeriesCrosstabDefaults = {
  ...MultiTimeSeriesDefaults,
  data: MultiTimeSeriesCrosstabDataSample,
};

export const MultiTimeSeriesGroupedDefaults = {
  data: MultiTimeSeriesDataSample,
  properties: {
    ...(MultiTimeSeriesDefaults.properties),
    grouping: "compare",
  },
};