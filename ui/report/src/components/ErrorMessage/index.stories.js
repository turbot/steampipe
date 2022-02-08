import ErrorMessage from "./index";

const story = {
  title: "Utilities/Error Message",
  component: ErrorMessage,
};

export default story;

const Template = (args) => <ErrorMessage {...args} />;

export const NoError = Template.bind({});
NoError.args = {
  error: null,
};

export const CustomFallback = Template.bind({});
CustomFallback.args = {
  error: {},
  fallbackMessage: "I am a fallback message!",
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
