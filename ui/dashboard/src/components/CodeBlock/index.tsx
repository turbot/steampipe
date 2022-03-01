import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { classNames } from "../../utils/styles";
import {
  vs,
  vscDarkPlus,
} from "react-syntax-highlighter/dist/esm/styles/prism";
import { ThemeNames, useTheme } from "../../hooks/useTheme";
import { useMemo } from "react";

interface CodeBlockProps {
  children: string;
  language?: "hcl" | "json" | "sql";
  onClick?: () => void;
  style?: any;
}

const CodeBlock = ({
  children,
  language = "sql",
  onClick,
  style = {},
}: CodeBlockProps) => {
  const { theme } = useTheme();
  const styles = useMemo(() => {
    if (theme.name === ThemeNames.STEAMPIPE_DARK) {
      return {
        ...vscDarkPlus,
        'code[class*="language-"]': {
          ...vscDarkPlus['code[class*="language-"]'],
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;',
          fontSize: "13px",
          lineHeight: 1.5,
          margin: 0,
        },
        'pre[class*="language-"]': {
          ...vscDarkPlus['pre[class*="language-"]'],
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;',
          fontSize: "13px",
          lineHeight: 1.5,
          margin: 0,
        },
      };
    } else {
      return {
        ...vs,
        'code[class*="language-"]': {
          ...vs['code[class*="language-"]'],
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;',
          fontSize: "13px",
          lineHeight: 1.5,
          margin: 0,
        },
        'pre > code[class*="language-"]': {
          ...vs['pre > code[class*="language-"]'],
          fontSize: "13px",
          lineHeight: 1.5,
          margin: 0,
        },
        'pre[class*="language-"]': {
          ...vs['pre[class*="language-"]'],
          border: "none",
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;',
          fontSize: "13px",
          lineHeight: 1.5,
          margin: 0,
        },
      };
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
          background: "transparent",
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
