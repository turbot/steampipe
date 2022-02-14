import Card from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";

const story = {
  title: "Primitives/Card",
  component: Card,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType="card" />
);

export const Loading = Template.bind({});
Loading.args = {
  data: null,
};

export const LoadingCustomIcon = Template.bind({});
LoadingCustomIcon.storyName = "Loading (Custom Icon)";
LoadingCustomIcon.args = {
  data: null,
  properties: { icon: "heroicons-solid:question-mark-circle" },
};

export const LoadingOK = Template.bind({});
LoadingOK.args = {
  data: null,
  properties: { type: "ok" },
};

export const LoadingOKCustomIcon = Template.bind({});
LoadingOKCustomIcon.storyName = "Loading OK (Custom Icon)";
LoadingOKCustomIcon.args = {
  data: null,
  properties: { type: "ok", icon: "check-circle" },
};

export const LoadingAlert = Template.bind({});
LoadingAlert.args = {
  data: null,
  properties: { type: "alert" },
};

export const LoadingAlertCustomIcon = Template.bind({});
LoadingAlertCustomIcon.storyName = "Loading Alert (Custom Icon)";
LoadingAlertCustomIcon.args = {
  data: null,
  properties: { type: "alert", icon: "shield-exclamation" },
};

export const LoadingInfo = Template.bind({});
LoadingInfo.args = {
  data: null,
  properties: { type: "info" },
};

export const LoadingInfoCustomIcon = Template.bind({});
LoadingInfoCustomIcon.storyName = "Loading Info (Custom Icon)";
LoadingInfoCustomIcon.args = {
  data: null,
  properties: { type: "info", icon: "light-bulb" },
};

export const Error = Template.bind({});
Error.args = {
  data: null,
  error: "Something went wrong!",
};

export const Empty = Template.bind({});
Empty.args = {
  data: [],
};

export const EmptyOK = Template.bind({});
EmptyOK.args = {
  data: [],
  properties: { type: "ok" },
};

export const EmptyAlert = Template.bind({});
EmptyAlert.args = {
  data: [],
  properties: { type: "alert" },
};

export const EmptyInfo = Template.bind({});
EmptyInfo.args = {
  data: [],
  properties: { type: "info" },
};

export const StringValue = Template.bind({});
StringValue.args = {
  data: {
    columns: [{ name: "Label", data_type_name: "INT8" }],
    rows: [["I am not a number"]],
  },
};

export const JSONValue = Template.bind({});
JSONValue.args = {
  data: {
    columns: [{ name: "Label", data_type_name: "INT8" }],
    rows: [[{ complex: "object" }]],
  },
};

export const SimpleDataFormat = Template.bind({});
SimpleDataFormat.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[106]],
  },
};

export const SimpleDataFormatOK = Template.bind({});
SimpleDataFormatOK.args = {
  data: {
    columns: [{ name: "Encrypted EC2 Instances", data_type_name: "INT8" }],
    rows: [[5]],
  },
  properties: { type: "ok" },
};

export const SimpleDataFormatOKCustomIcon = Template.bind({});
SimpleDataFormatOKCustomIcon.storyName = "Simple Data Format OK (Custom Icon)";
SimpleDataFormatOKCustomIcon.args = {
  data: {
    columns: [{ name: "Encrypted EC2 Instances", data_type_name: "INT8" }],
    rows: [[5]],
  },
  properties: { type: "ok", icon: "check-circle" },
};

export const SimpleDataFormatAlert = Template.bind({});
SimpleDataFormatAlert.args = {
  data: {
    columns: [{ name: "Public Buckets", data_type_name: "INT8" }],
    rows: [[5]],
  },
  properties: { type: "alert" },
};

export const SimpleDataFormatAlertCustomIcon = Template.bind({});
SimpleDataFormatAlertCustomIcon.storyName =
  "Simple Data Format Alert (Custom Icon)";
SimpleDataFormatAlertCustomIcon.args = {
  data: {
    columns: [{ name: "Public Buckets", data_type_name: "INT8" }],
    rows: [[5]],
  },
  properties: { type: "alert", icon: "shield-exclamation" },
};

export const SimpleDataFormatInfo = Template.bind({});
SimpleDataFormatInfo.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[106]],
  },
  properties: { type: "info" },
};

export const SimpleDataFormatInfoCustomIcon = Template.bind({});
SimpleDataFormatInfoCustomIcon.storyName =
  "Simple Data Format Info (Custom Icon)";
SimpleDataFormatInfoCustomIcon.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[106]],
  },
  properties: { type: "info", icon: "light-bulb" },
};

export const SimpleDataFormatThousands = Template.bind({});
SimpleDataFormatThousands.storyName = "Simple Data Format (thousands)";
SimpleDataFormatThousands.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type_name: "INT8" }],
    rows: [[1236]],
  },
  properties: { type: "info" },
};

export const SimpleDataFormatMillions = Template.bind({});
SimpleDataFormatMillions.storyName = "Simple Data Format (millions)";
SimpleDataFormatMillions.args = {
  data: {
    columns: [{ name: "Log Lines", data_type_name: "INT8" }],
    rows: [[5236174]],
  },
  properties: { type: "info" },
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

export const FormalDataFormatOK = Template.bind({});
FormalDataFormatOK.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["Encrypted EC2 Instances", 5, "ok"]],
  },
};

export const FormalDataFormatOKCustomIcon = Template.bind({});
FormalDataFormatOKCustomIcon.storyName = "Formal Data Format OK (Custom Icon)";
FormalDataFormatOKCustomIcon.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["Encrypted EC2 Instances", 5, "ok"]],
  },
  properties: { icon: "check-circle" },
};

export const FormalDataFormatOKCustomIconFromSQL = Template.bind({});
FormalDataFormatOKCustomIconFromSQL.storyName =
  "Formal Data Format OK (Custom Icon from SQL)";
FormalDataFormatOKCustomIconFromSQL.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
      { name: "icon", data_type_name: "TEXT" },
    ],
    rows: [
      ["Encrypted EC2 Instances", 5, "ok", "heroicons-solid:check-circle"],
    ],
  },
  properties: { icon: "check-circle" },
};

export const FormalDataFormatAlert = Template.bind({});
FormalDataFormatAlert.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["Public Buckets", 5, "alert"]],
  },
};

export const FormalDataFormatAlertCustomIcon = Template.bind({});
FormalDataFormatAlertCustomIcon.storyName =
  "Formal Data Format Alert (Custom Icon)";
FormalDataFormatAlertCustomIcon.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["Public Buckets", 5, "alert"]],
  },
  properties: { icon: "shield-exclamation" },
};

export const FormalDataFormatAlertCustomIconFromSQL = Template.bind({});
FormalDataFormatAlertCustomIconFromSQL.storyName =
  "Formal Data Format Alert (Custom Icon from SQL)";
FormalDataFormatAlertCustomIconFromSQL.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
      { name: "icon", data_type_name: "TEXT" },
    ],
    rows: [
      ["Public Buckets", 5, "alert", "heroicons-solid:shield-exclamation"],
    ],
  },
  properties: { icon: "shield-exclamation" },
};

export const FormalDataFormatInfo = Template.bind({});
FormalDataFormatInfo.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["EC2 Instances", 106, "info"]],
  },
};

export const FormalDataFormatInfoCustomIcon = Template.bind({});
FormalDataFormatInfoCustomIcon.storyName =
  "Formal Data Format Info (Custom Icon)";
FormalDataFormatInfoCustomIcon.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["EC2 Instances", 106, "info"]],
  },
  properties: { icon: "light-bulb" },
};

export const FormalDataFormatInfoCustomIconFromSQL = Template.bind({});
FormalDataFormatInfoCustomIconFromSQL.storyName =
  "Formal Data Format Info (Custom Icon from SQL)";
FormalDataFormatInfoCustomIconFromSQL.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
      { name: "icon", data_type_name: "TEXT" },
    ],
    rows: [["EC2 Instances", 106, "info", "heroicons-solid:light-bulb"]],
  },
  properties: { icon: "light-bulb" },
};

export const FormalDataFormatAsTable = Template.bind({});
FormalDataFormatAsTable.args = {
  data: {
    columns: [
      { name: "label", data_type_name: "TEXT" },
      { name: "value", data_type_name: "INT8" },
      { name: "type", data_type_name: "TEXT" },
    ],
    rows: [["Encrypted EC2 Instances", 5, "ok"]],
  },
  properties: {
    type: "table",
  },
};
