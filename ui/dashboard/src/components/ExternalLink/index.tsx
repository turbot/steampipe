import { Link } from "react-router-dom";

interface ExternalLinkProps {
  children: null | JSX.Element | JSX.Element[];
  className?: string;
  target?: string;
  title?: string;
  to: string;
  withReferrer?: boolean;
}

const ExternalLink = ({
  children,
  className = "link-highlight",
  target = "_blank",
  title,
  to,
  withReferrer = false,
}: ExternalLinkProps) => {
  if (!to) {
    return null;
  }

  if (to.match("^https?://")) {
    return (
      /*eslint-disable */
      <a
        className={className}
        href={to}
        rel={withReferrer ? undefined : "noopener noreferrer"}
        target={target}
      >
        {children}
      </a>
      /*eslint-enable */
    );
  }

  return (
    <Link to={to} className={className} title={title}>
      {children}
    </Link>
  );
};

export default ExternalLink;
