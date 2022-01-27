import Chart from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Charts/Chart",
  component: Chart,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType="chart" />
);

export const DefaultsToColumn = Template.bind({});
DefaultsToColumn.args = {
  data: {
    columns: [
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    items: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
};
