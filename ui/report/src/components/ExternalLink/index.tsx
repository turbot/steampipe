const ExternalLink = ({
  children,
  className = "link-highlight",
  url,
  withReferrer = false,
}) => (
  <a
    className={className}
    href={url}
    rel={withReferrer ? undefined : "noopener noreferrer"}
    target="_blank"
  >
    {children}
  </a>
);

export default ExternalLink;
