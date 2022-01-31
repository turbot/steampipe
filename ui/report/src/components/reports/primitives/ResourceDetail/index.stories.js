import ResourceDetail from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Primitives/Resource Detail",
  component: ResourceDetail.component,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType={ResourceDetail.type} />
);

export const Loading = Template.bind({});
Loading.args = {
  data: null,
};

export const NoData = Template.bind({});
NoData.args = {
  data: [["Name", "VPC", "Region"]],
};

export const Basic = Template.bind({});
Basic.args = {
  data: [
    ["Name", "VPC", "Region"],
    ["smyth-test-sg1", "vpc-0bf2ca1f6a9319eea [172.16.0.0/16]", "us-east-1"],
  ],
};
