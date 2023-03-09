import Card from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";

const story = {
  title: "Primitives/Card",
  component: Card,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} panelType="card" />
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
  display_type: "ok",
};

export const LoadingOKCustomIcon = Template.bind({});
LoadingOKCustomIcon.storyName = "Loading OK (Custom Icon)";
LoadingOKCustomIcon.args = {
  data: null,
  display_type: "ok",
  properties: { icon: "check-circle" },
};

export const LoadingAlert = Template.bind({});
LoadingAlert.args = {
  data: null,
  display_type: "alert",
};

export const LoadingAlertCustomIcon = Template.bind({});
LoadingAlertCustomIcon.storyName = "Loading Alert (Custom Icon)";
LoadingAlertCustomIcon.args = {
  data: null,
  display_type: "alert",
  properties: { icon: "shield-exclamation" },
};

export const LoadingInfo = Template.bind({});
LoadingInfo.args = {
  data: null,
  display_type: "info",
};

export const LoadingInfoCustomIcon = Template.bind({});
LoadingInfoCustomIcon.storyName = "Loading Info (Custom Icon)";
LoadingInfoCustomIcon.args = {
  data: null,
  display_type: "info",
  properties: { icon: "light-bulb" },
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
  display_type: "ok",
};

export const EmptyAlert = Template.bind({});
EmptyAlert.args = {
  data: [],
  display_type: "alert",
};

export const EmptyInfo = Template.bind({});
EmptyInfo.args = {
  data: [],
  display_type: "info",
};

export const StringValue = Template.bind({});
StringValue.args = {
  data: {
    columns: [{ name: "Label", data_type: "TEXT" }],
    rows: [{ Label: "I am not a number" }],
  },
};

export const JSONValue = Template.bind({});
JSONValue.args = {
  data: {
    columns: [{ name: "Label", data_type: "JSONB" }],
    rows: [{ Label: { complex: "object" } }],
  },
};

export const SimpleDataFormat = Template.bind({});
SimpleDataFormat.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type: "INT8" }],
    rows: [{ "EC2 Instances": 106 }],
  },
};

export const SimpleDataFormatOK = Template.bind({});
SimpleDataFormatOK.args = {
  data: {
    columns: [{ name: "Encrypted EC2 Instances", data_type: "INT8" }],
    rows: [{ "Encrypted EC2 Instances": 5 }],
  },
  display_type: "ok",
};

export const SimpleDataFormatOKCustomIcon = Template.bind({});
SimpleDataFormatOKCustomIcon.storyName = "Simple Data Format OK (Custom Icon)";
SimpleDataFormatOKCustomIcon.args = {
  data: {
    columns: [{ name: "Encrypted EC2 Instances", data_type: "INT8" }],
    rows: [{ "Encrypted EC2 Instances": 5 }],
  },
  display_type: "ok",
  properties: { icon: "check-circle" },
};

export const SimpleDataFormatAlert = Template.bind({});
SimpleDataFormatAlert.args = {
  data: {
    columns: [{ name: "Public Buckets", data_type: "INT8" }],
    rows: [{ "Public Buckets": 5 }],
  },
  display_type: "alert",
};

export const SimpleDataFormatAlertCustomIcon = Template.bind({});
SimpleDataFormatAlertCustomIcon.storyName =
  "Simple Data Format Alert (Custom Icon)";
SimpleDataFormatAlertCustomIcon.args = {
  data: {
    columns: [{ name: "Public Buckets", data_type: "INT8" }],
    rows: [{ "Public Buckets": 5 }],
  },
  display_type: "alert",
  properties: { icon: "shield-exclamation" },
};

export const SimpleDataFormatInfo = Template.bind({});
SimpleDataFormatInfo.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type: "INT8" }],
    rows: [{ "EC2 Instances": 106 }],
  },
  display_type: "info",
};

export const SimpleDataFormatInfoCustomIcon = Template.bind({});
SimpleDataFormatInfoCustomIcon.storyName =
  "Simple Data Format Info (Custom Icon)";
SimpleDataFormatInfoCustomIcon.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type: "INT8" }],
    rows: [{ "EC2 Instances": 106 }],
  },
  display_type: "info",
  properties: { icon: "light-bulb" },
};

export const SimpleDataFormatThousands = Template.bind({});
SimpleDataFormatThousands.storyName = "Simple Data Format (thousands)";
SimpleDataFormatThousands.args = {
  data: {
    columns: [{ name: "EC2 Instances", data_type: "INT8" }],
    rows: [{ "EC2 Instances": 1236 }],
  },
  display_type: "info",
};

export const SimpleDataFormatMillions = Template.bind({});
SimpleDataFormatMillions.storyName = "Simple Data Format (millions)";
SimpleDataFormatMillions.args = {
  data: {
    columns: [{ name: "Log Lines", data_type: "INT8" }],
    rows: [{ "Log Lines": 5236174 }],
  },
  display_type: "info",
};

export const FormalDataFormat = Template.bind({});
FormalDataFormat.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
    ],
    rows: [{ label: "EC2 Instances", value: 106 }],
  },
};

export const FormalDataFormatOK = Template.bind({});
FormalDataFormatOK.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [{ label: "Encrypted EC2 Instances", value: 5, type: "ok" }],
  },
};

export const FormalDataFormatOKCustomIcon = Template.bind({});
FormalDataFormatOKCustomIcon.storyName = "Formal Data Format OK (Custom Icon)";
FormalDataFormatOKCustomIcon.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [{ label: "Encrypted EC2 Instances", value: 5, type: "ok" }],
  },
  properties: { icon: "check-circle" },
};

export const FormalDataFormatOKCustomIconFromSQL = Template.bind({});
FormalDataFormatOKCustomIconFromSQL.storyName =
  "Formal Data Format OK (Custom Icon from SQL)";
FormalDataFormatOKCustomIconFromSQL.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
      { name: "icon", data_type: "TEXT" },
    ],
    rows: [
      {
        label: "Encrypted EC2 Instances",
        value: 5,
        type: "ok",
        icon: "heroicons-solid:check-circle",
      },
    ],
  },
  properties: { icon: "check-circle" },
};

export const FormalDataFormatAlert = Template.bind({});
FormalDataFormatAlert.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [{ label: "Public Buckets", value: 5, type: "alert" }],
  },
};

export const FormalDataFormatAlertCustomIcon = Template.bind({});
FormalDataFormatAlertCustomIcon.storyName =
  "Formal Data Format Alert (Custom Icon)";
FormalDataFormatAlertCustomIcon.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [{ label: "Public Buckets", value: 5, type: "alert" }],
  },
  properties: { icon: "shield-exclamation" },
};

export const FormalDataFormatAlertCustomIconFromSQL = Template.bind({});
FormalDataFormatAlertCustomIconFromSQL.storyName =
  "Formal Data Format Alert (Custom Icon from SQL)";
FormalDataFormatAlertCustomIconFromSQL.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
      { name: "icon", data_type: "TEXT" },
    ],
    rows: [
      {
        label: "Public Buckets",
        value: 5,
        type: "alert",
        icon: "heroicons-solid:shield-exclamation",
      },
    ],
  },
  properties: { icon: "shield-exclamation" },
};

export const FormalDataFormatInfo = Template.bind({});
FormalDataFormatInfo.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [{ label: "EC2 Instances", value: 106, type: "info" }],
  },
};

export const FormalDataFormatInfoCustomIcon = Template.bind({});
FormalDataFormatInfoCustomIcon.storyName =
  "Formal Data Format Info (Custom Icon)";
FormalDataFormatInfoCustomIcon.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [{ label: "EC2 Instances", value: 106, type: "info" }],
  },
  properties: { icon: "light-bulb" },
};

export const FormalDataFormatInfoCustomIconFromSQL = Template.bind({});
FormalDataFormatInfoCustomIconFromSQL.storyName =
  "Formal Data Format Info (Custom Icon from SQL)";
FormalDataFormatInfoCustomIconFromSQL.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
      { name: "icon", data_type: "TEXT" },
    ],
    rows: [
      {
        label: "EC2 Instances",
        value: 106,
        type: "info",
        icon: "heroicons-solid:light-bulb",
      },
    ],
  },
  properties: { icon: "light-bulb" },
};

export const FormalDataFormatAsTable = Template.bind({});
FormalDataFormatAsTable.args = {
  data: {
    columns: [
      { name: "label", data_type: "TEXT" },
      { name: "value", data_type: "INT8" },
      { name: "type", data_type: "TEXT" },
    ],
    rows: [
      {
        label: "Encrypted EC2 Instances",
        value: 5,
        type: "ok",
      },
    ],
  },
  properties: {
    type: "table",
  },
};
