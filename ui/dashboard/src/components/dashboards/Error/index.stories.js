import Error from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";

const story = {
  title: "Utilities/Error",
  component: Error,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} panelType="error" />
);

export const NoError = Template.bind({});
NoError.args = {
  error: null,
};

export const StringError = Template.bind({});
StringError.args = {
  error: "Something went wrong!",
};

export const ErrorObjectLowerCaseErrorMessage = Template.bind({});
ErrorObjectLowerCaseErrorMessage.args = {
  error: { message: "Something went wrong!" },
};

export const ErrorObjectUpperCaseErrorMessage = Template.bind({});
ErrorObjectUpperCaseErrorMessage.args = {
  error: { Message: "Something went wrong!" },
};

export const ErrorObjectNoErrorMessage = Template.bind({});
ErrorObjectNoErrorMessage.args = {
  error: { some: "value" },
};

export const UnknownErrorObject = Template.bind({});
UnknownErrorObject.args = {
  error: 12,
};
