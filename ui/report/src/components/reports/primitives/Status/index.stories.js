import Status from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Primitives/Status",
  component: Status.component,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType={Status.type} />
);

export const Loading = Template.bind({});
Loading.args = {
  loading: true,
};

export const Error = Template.bind({});
Error.args = {
  loading: false,
  error: "Something went wrong!",
};

export const Basic = Template.bind({});
Basic.args = {
  data: [["EC2 Instances"], [106]],
};

export const Info = Template.bind({});
Info.args = {
  data: [["EC2 Instances"], [106]],
  type: "info",
};

export const Alert = Template.bind({});
Alert.args = {
  data: [["Public Buckets"], [5]],
  type: "alert",
};

export const Thousands = Template.bind({});
Thousands.args = {
  data: [["EC2 Instances"], [1236]],
};

export const Millions = Template.bind({});
Millions.args = {
  data: [["Log Lines"], [5236174]],
};
