import Counter from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";

const story = {
  title: "Primitives/Counter",
  component: Counter,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType="counter" />
);

export const Loading = Template.bind({});
Loading.args = {
  data: null,
  properties: { style: "info" },
};

export const Error = Template.bind({});
Error.args = {
  data: null,
  error: "Something went wrong!",
};

export const Empty = Template.bind({});
Empty.args = {
  data: [],
  properties: { style: "info" },
};

export const SimpleDataFormat = Template.bind({});
SimpleDataFormat.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[106]],
  },
};

export const SimpleDataFormatInfo = Template.bind({});
SimpleDataFormatInfo.storyName = "Simple Data Format (info)";
SimpleDataFormatInfo.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[106]],
  },
  properties: { style: "info" },
};

export const SimpleDataFormatAlert = Template.bind({});
SimpleDataFormatAlert.storyName = "Simple Data Format (alert)";
SimpleDataFormatAlert.args = {
  data: {
    columns: [{ name: "Public Buckets", data_type_name: "INT8" }],
    rows: [[5]],
  },
  properties: { style: "alert" },
};

export const SimpleDataFormatOK = Template.bind({});
SimpleDataFormatOK.storyName = "Simple Data Format (ok)";
SimpleDataFormatOK.args = {
  data: {
    columns: [{ name: "Encrypted EC2 Instances", data_type_name: "INT8" }],
    rows: [[5]],
  },
  properties: { style: "ok" },
};

export const SimpleDataFormatThousands = Template.bind({});
SimpleDataFormatThousands.storyName = "Simple Data Format (thousands)";
SimpleDataFormatThousands.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[1236]],
  },
  properties: { style: "info" },
};

export const SimpleDataFormatMillions = Template.bind({});
SimpleDataFormatMillions.storyName = "Simple Data Format (millions)";
SimpleDataFormatMillions.args = {
  data: {
    columns: [{ name: "Log Lines", data_type_name: "INT8" }],
    rows: [[5236174]],
  },
  properties: { style: "info" },
};

export const FormalDataFormat = Template.bind({});
FormalDataFormat.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
    ],
    rows: [["EC2 Instances", 106]],
  },
};

export const FormalDataFormatInfo = Template.bind({});
FormalDataFormatInfo.storyName = "Formal Data Format (info)";
FormalDataFormatInfo.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "style", data_type_name: "TEXT" },
    ],
    rows: [["EC2 Instances", 106, "info"]],
  },
};

export const FormalDataFormatAlert = Template.bind({});
FormalDataFormatAlert.storyName = "Formal Data Format (alert)";
FormalDataFormatAlert.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "style", data_type_name: "TEXT" },
    ],
    rows: [["Public Buckets", 5, "alert"]],
  },
};

export const FormalDataFormatOK = Template.bind({});
FormalDataFormatOK.storyName = "Formal Data Format (ok)";
FormalDataFormatOK.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "style", data_type_name: "TEXT" },
    ],
    rows: [["Encrypted EC2 Instances", 5, "ok"]],
  },
};

export const FormalDataFormatAsTable = Template.bind({});
FormalDataFormatAsTable.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "style", data_type_name: "TEXT" },
    ],
    rows: [["Encrypted EC2 Instances", 5, "ok"]],
  },
  properties: {
    type: "table",
  },
};
