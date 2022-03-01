import Container from "../layout/Container";
import { PanelDefinition } from "../../../hooks/useDashboard";

type TitleWrapperProps = {
  children: null | JSX.Element | JSX.Element[];
  definition: PanelDefinition;
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
  const innerProperties = definitionOther.properties || {};
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
        original_title: definitionTitle,
        properties: {
          ...innerProperties,
          parentWidth: definitionWidth,
        },
      },
    ],
  };
  return <Container definition={wrappedDefinition} withNarrowVertical={true} />;
};

export default TitleWrapper;
