import ResourceTags from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Primitives/Resource Tags",
  component: ResourceTags.component,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType={ResourceTags.type} />
);

export const Loading = Template.bind({});
Loading.args = {
  data: null,
};

export const NoData = Template.bind({});
NoData.args = {
  data: [["Tags"]],
};

export const Basic = Template.bind({});
Basic.args = {
  data: [["Tags"], [{ cost_center: "z-1234", my_key: "value", owner: "Bob" }]],
};
