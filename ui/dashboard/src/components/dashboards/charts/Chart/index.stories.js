import Chart from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Charts/Chart",
  component: Chart,
  excludeStories: ["SingleTimeSeriesDefaults", "MultiTimeSeriesDefaults", "MultiTimeSeriesGroupedDefaults"]
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
      { name: "time", data_type: "TEXT" },
      { name: "count", data_type: "INT8" },
    ],
    rows: [
      { time: "2023-01", count: 20 },
      { time: "2023-02", count: 32 },
      { time: "2023-04", count: -15 },
      { time: "2023-05", count: 18 },
      { time: "2023-06", count: -9 },
      { time: "2023-12", count: 3 },
    ],
  },
  properties: {
    axes: {
      x: {
        type: "time",
      }
    },
  },
};

const MultiTimeSeriesDataSample = {
  columns: [
    { name: "time", data_type: "TEXT" },
    { name: "Income", data_type: "INT8" },
    { name: "Spending", data_type: "INT8" },
  ],
  rows: [
    { time: "2023-01", Income: 20, Spending: 0 },
    { time: "2023-02", Income: 18, Spending: 32 },
    { time: "2023-04", Income: 15, Spending: 3 },
    { time: "2023-05", Income: 18, Spending: 15 },
    { time: "2023-06", Income: 0, Spending: 9 },
    { time: "2023-09", Income: 7, Spending: 26 },
    { time: "2023-12", Income: 8, Spending: 3 },
  ],
};

export const MultiTimeSeriesDefaults = {
  data: MultiTimeSeriesDataSample,
  properties: {
    axes: {
      x: {
        type: "time",
      }
    },
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

export const MultiTimeSeriesGroupedDefaults = {
  data: MultiTimeSeriesDataSample,
  properties: {
    ...(MultiTimeSeriesDefaults.properties),
    grouping: "compare",
  },
};