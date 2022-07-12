import Error from "../Error";
import gfm from "remark-gfm"; // Support for strikethrough, tables, tasklists and URLs
import ReactMarkdown from "react-markdown";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { classNames } from "../../../utils/styles";
import { registerComponent } from "../index";

const getLongPanelClasses = () => {
  // switch (type) {
  // case "alert":
  //   return "p-2 border border-alert bg-alert-light border overflow-hidden sm:rounded-md";
  // default:
  return "overflow-hidden sm:rounded-md";
  // }
};

const getShortPanelClasses = () => {
  // switch (type) {
  //   case "alert":
  //     return "p-2 border border-alert bg-alert-light prose prose-sm sm:rounded-md max-w-none";
  //   default:
  return "prose prose-sm sm:rounded-md max-w-none";
  // }
};

export type TextProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    display_type?: "raw" | "markdown" | "html";
    properties: {
      value: string;
    };
  };

const Markdown = ({ value }) => {
  if (!value) {
    return null;
  }

  // Use standard prose styles from Tailwind
  // Do not restrict width, that's the job of the wrapping panel
  const isLong = value.split("\n").length > 3;
  const panelClasses = isLong ? getLongPanelClasses() : getShortPanelClasses();
  const proseHeadings =
    "prose-h1:text-3xl prose-h2:text-2xl prose-h3:text-xl prose-h3:mt-1 p-4";

  return (
    <>
      {isLong ? (
        <div className={panelClasses}>
          <div
            className={classNames(
              "p-2 sm:p-1 prose prose-sm max-w-none break-all",
              proseHeadings
            )}
          >
            <ReactMarkdown remarkPlugins={[gfm]}>{value}</ReactMarkdown>
          </div>
        </div>
      ) : (
        <article
          className={classNames(panelClasses, "break-all", proseHeadings)}
        >
          <ReactMarkdown remarkPlugins={[gfm]}>{value}</ReactMarkdown>
        </article>
      )}
    </>
  );
};

const Raw = ({ value }) => {
  if (!value) {
    return null;
  }
  return <pre className="whitespace-pre-wrap break-all">{value}</pre>;
};

const renderText = (type, value) => {
  switch (type) {
    case "markdown":
      return <Markdown value={value} />;
    case "raw":
      return <Raw value={value} />;
    default:
      return <Error error={`Unsupported text type ${type}`} />;
  }
};

const Text = (props: TextProps) =>
  renderText(
    props.display_type || "markdown",
    props.properties ? props.properties.value : null
  );

registerComponent("text", Text);

export default Text;
