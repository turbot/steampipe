import { Link as ReactRouterLink } from "react-router-dom";

interface LinkProps {
  children: null | JSX.Element | JSX.Element[];
  link_url: string;
}

const Link = ({ children, link_url }: LinkProps) => {
  if (!link_url.startsWith("/")) {
    return (
      <a
        className="link-highlight"
        href={link_url}
        target="_blank"
        rel="noopener noreferrer"
      >
        {children}
      </a>
    );
  }
  return (
    <ReactRouterLink className="link-highlight" to={link_url}>
      {children}
    </ReactRouterLink>
  );
};

export default Link;
