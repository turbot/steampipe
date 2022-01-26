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
    items: [{ "EC2 Instances": 106 }],
  },
};

export const SimpleDataFormatInfo = Template.bind({});
SimpleDataFormatInfo.storyName = "Simple Data Format (info)";
SimpleDataFormatInfo.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    items: [{ "EC2 Instances": 106 }],
  },
  properties: { style: "info" },
};

export const SimpleDataFormatAlert = Template.bind({});
SimpleDataFormatAlert.storyName = "Simple Data Format (alert)";
SimpleDataFormatAlert.args = {
  data: {
    columns: [{ name: "Public Buckets", data_type_name: "INT8" }],
    items: [{ "Public Buckets": 5 }],
  },
  properties: { style: "alert" },
};

export const SimpleDataFormatOK = Template.bind({});
SimpleDataFormatOK.storyName = "Simple Data Format (ok)";
SimpleDataFormatOK.args = {
  data: {
    columns: [{ name: "Encrypted EC2 Instances", data_type_name: "INT8" }],
    items: [{ "Encrypted EC2 Instances": 5 }],
  },
  properties: { style: "ok" },
};

export const SimpleDataFormatThousands = Template.bind({});
SimpleDataFormatThousands.storyName = "Simple Data Format (thousands)";
SimpleDataFormatThousands.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    items: [{ "EC2 Instances": 1236 }],
  },
  properties: { style: "info" },
};

export const SimpleDataFormatMillions = Template.bind({});
SimpleDataFormatMillions.storyName = "Simple Data Format (millions)";
SimpleDataFormatMillions.args = {
  data: {
    columns: [{ name: "Log Lines", data_type_name: "INT8" }],
    items: [{ "Log Lines": 5236174 }],
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
    items: [{ label: "EC2 Instances", value: 106 }],
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
    items: [{ label: "EC2 Instances", value: 106, style: "info" }],
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
    items: [{ label: "Public Buckets", value: 5, style: "alert" }],
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
    items: [{ label: "Encrypted EC2 Instances", value: 5, style: "ok" }],
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
    items: [{ label: "Encrypted EC2 Instances", value: 5, style: "ok" }],
  },
  properties: {
    type: "table",
  },
};
