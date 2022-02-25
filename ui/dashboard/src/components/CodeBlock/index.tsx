import SyntaxHighlighter from "react-syntax-highlighter";
import { classNames } from "../../utils/styles";
import {
  dark,
  github,
  githubGist,
} from "react-syntax-highlighter/dist/esm/styles/hljs";
import { ThemeNames, useTheme } from "../../hooks/useTheme";
import { useMemo } from "react";

interface CodeBlockProps {
  children: string;
  language?: "hcl" | "json" | "sql";
  onClick?: () => void;
  style?: any;
  withBackground?: boolean;
}

const CodeBlock = ({
  children,
  language = "sql",
  onClick,
  style = {},
  withBackground = false,
}: CodeBlockProps) => {
  const { theme } = useTheme();
  const styles = useMemo(() => {
    if (language === "sql" && theme.name === ThemeNames.STEAMPIPE_DARK) {
      return {
        ...dark,
        "hljs-keyword": {
          ...dark["hljs-keyword"],
          color: "#0000FF",
        },
      };
    } else if (language === "sql" && theme.name !== ThemeNames.STEAMPIPE_DARK) {
      return {
        ...githubGist,
        "hljs-keyword": {
          ...githubGist["hljs-keyword"],
          color: "#0000FF",
        },
      };
    } else {
      return github;
    }
  }, [language, theme.name]);

  return (
    <div
      className={classNames(onClick ? "cursor-pointer" : "")}
      onClick={onClick}
    >
      <SyntaxHighlighter
        language={language}
        style={styles}
        customStyle={{
          padding: 0,
          wordBreak: "break-all",
          background:
            withBackground && theme.name === ThemeNames.STEAMPIPE_DARK
              ? "var(--color-black-scale-2)"
              : withBackground
              ? "var(--color-black-scale-2)"
              : "transparent",
          borderRadius: "4px",
          ...style,
        }}
        wrapLongLines
      >
        {children || ""}
      </SyntaxHighlighter>
    </div>
  );
};

export default CodeBlock;
