import { Link } from "react-router-dom";
import {
  DashboardDataModeCLISnapshot,
  DashboardDataModeCloudSnapshot,
} from "../../types";
import { ReactNode } from "react";
import { registerComponent } from "../dashboards";
import { useDashboard } from "../../hooks/useDashboard";

type ExternalLinkProps = {
  children: ReactNode;
  className?: string;
  ignoreDataMode?: boolean;
  target?: string;
  title?: string;
  to: string;
  withReferrer?: boolean;
};

const ExternalLink = ({
  children,
  className = "link-highlight",
  ignoreDataMode = false,
  target = "_blank",
  title,
  to,
  withReferrer = false,
}: ExternalLinkProps) => {
  const { dataMode } = useDashboard();

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

  if (
    (!ignoreDataMode && dataMode === DashboardDataModeCLISnapshot) ||
    dataMode === DashboardDataModeCloudSnapshot
  ) {
    return children || null;
  }

  return (
    <Link to={to} className={className} title={title}>
      {children}
    </Link>
  );
};

registerComponent("external_link", ExternalLink);

export default ExternalLink;
