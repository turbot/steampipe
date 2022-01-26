import gfm from "remark-gfm"; // Support for strikethrough, tables, tasklists and URLs
import ReactMarkdown from "react-markdown";
import { BasePrimitiveProps } from "../common";

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

export type TextProps = BasePrimitiveProps & {
  properties: {
    type?: "raw" | "markdown" | "html";
    value: string;
  };
};

const Markdown = ({ value }) => {
  // Use standard prose styles from Tailwind
  // Do not restrict width, that's the job of the wrapping panel
  const isLong = value.split("\n").length > 3;
  const panelClasses = isLong ? getLongPanelClasses() : getShortPanelClasses();

  return (
    <>
      {isLong ? (
        <div className={panelClasses}>
          <div className="p-2 sm:p-1 prose prose-sm max-w-none">
            <ReactMarkdown remarkPlugins={[gfm]}>{value}</ReactMarkdown>
          </div>
        </div>
      ) : (
        <article className={panelClasses}>
          <ReactMarkdown remarkPlugins={[gfm]}>{value}</ReactMarkdown>
        </article>
      )}
    </>
  );
};

const Text = (props: TextProps) => {
  const type = props.properties.type ? props.properties.type : "markdown";
  return (
    <>
      {type === "raw" && <>{props.properties.value}</>}
      {type === "markdown" && <Markdown value={props.properties.value} />}
    </>
  );
};

export default Text;
