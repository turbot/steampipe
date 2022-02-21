const ExternalLink = ({
  children,
  className = "link-highlight",
  url,
  withReferrer = false,
}) => (
  /*eslint-disable */
  <a
    className={className}
    href={url}
    rel={withReferrer ? undefined : "noopener noreferrer"}
    target="_blank"
  >
    {children}
  </a>
  /*eslint-enable */
);

export default ExternalLink;
