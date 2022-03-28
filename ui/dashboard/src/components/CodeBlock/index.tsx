import CopyToClipboard, { CopyToClipboardProvider } from "../CopyToClipboard";
import { classNames } from "../../utils/styles";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { ThemeNames } from "../../hooks/useTheme";
import { useDashboard } from "../../hooks/useDashboard";
import { useMemo, useState } from "react";
import {
  vs,
  vscDarkPlus,
} from "react-syntax-highlighter/dist/esm/styles/prism";

interface CodeBlockProps {
  children: string;
  copyToClipboard?: boolean;
  language?: "hcl" | "json" | "sql";
  style?: any;
}

const CodeBlock = ({
  children,
  copyToClipboard = true,
  language = "sql",
  style = {},
}: CodeBlockProps) => {
  const [showCopyIcon, setShowCopyIcon] = useState(false);
  const {
    themeContext: { theme },
  } = useDashboard();

  const styles = useMemo(() => {
    const commonStyles = {
      fontFamily:
        'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
      fontSize: "13px",
      lineHeight: 1.5,
      margin: 0,
    };
    if (theme.name === ThemeNames.STEAMPIPE_DARK) {
      return {
        ...vscDarkPlus,
        'code[class*="language-"]': {
          ...vscDarkPlus['code[class*="language-"]'],
          ...commonStyles,
        },
        'pre[class*="language-"]': {
          ...vscDarkPlus['pre[class*="language-"]'],
          ...commonStyles,
        },
      };
    } else {
      return {
        ...vs,
        'code[class*="language-"]': {
          ...vs['code[class*="language-"]'],
          ...commonStyles,
        },
        'pre > code[class*="language-"]': {
          ...vs['pre > code[class*="language-"]'],
          ...commonStyles,
        },
        'pre[class*="language-"]': {
          ...vs['pre[class*="language-"]'],
          border: "none",
          ...commonStyles,
        },
      };
    }
  }, [theme.name]);

  return (
    <CopyToClipboardProvider>
      {({ setDoCopy }) => (
        <div
          className={classNames(
            "relative p-1",
            copyToClipboard ? "cursor-pointer" : null,
            copyToClipboard && showCopyIcon ? "bg-black-scale-1" : null
          )}
          onMouseEnter={
            copyToClipboard
              ? () => {
                  setShowCopyIcon(true);
                }
              : undefined
          }
          onMouseLeave={
            copyToClipboard
              ? () => {
                  setShowCopyIcon(false);
                }
              : undefined
          }
          onClick={() => setDoCopy(true)}
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
          {showCopyIcon && (
            <div
              className={classNames(
                "absolute cursor-pointer z-50 right-1 top-1"
              )}
            >
              <CopyToClipboard data={children} />
            </div>
          )}
        </div>
      )}
    </CopyToClipboardProvider>
  );
};

export default CodeBlock;
