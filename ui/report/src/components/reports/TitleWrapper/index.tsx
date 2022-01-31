import Container from "../layout/Container";
import {
  ContainerDefinition,
  ReportDefinition,
} from "../../../hooks/useReport";
import { TextProps } from "../Text";

type TitleWrapperProps = {
  children: null | JSX.Element | JSX.Element[];
  definition: ReportDefinition | ContainerDefinition | TextProps;
  level: "container" | "panel";
  title: string;
};

const getMarkdownTitleLevel = (level) => {
  switch (level) {
    case "container":
      return "##";
    default:
      return "### ";
  }
};

const TitleWrapper = ({
  children,
  definition,
  level,
  title,
}: TitleWrapperProps) => {
  if (!title) {
    return <>{children}</>;
  }
  const markdownTitle = `${getMarkdownTitleLevel(level)} ${title}`;
  // We want to remove the original title and width here as leaving the title
  // will cause an infinite render loop and we apply the width to the wrapping
  // container, leaving the wrapped primitive full-width
  const {
    title: definitionTitle,
    width: definitionWidth,
    ...definitionOther
  } = definition;
  const wrappedDefinition = {
    name: `${definition.name}.container.wrapper`,
    width: definitionWidth,
    children: [
      {
        name: `${definition.name}.text.title`,
        node_type: "text",
        properties: {
          type: "markdown",
          value: markdownTitle,
        },
      },
      {
        ...definitionOther,
      },
    ],
  };
  return <Container definition={wrappedDefinition} />;
};

export default TitleWrapper;
