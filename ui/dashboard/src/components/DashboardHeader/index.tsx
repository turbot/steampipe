import ExternalLink from "../ExternalLink";
import ThemeToggle from "../ThemeToggle";
import { Link } from "react-router-dom";

const DashboardHeader = () => (
  <div className="flex w-screen p-3 items-center bg-background border-b border-black-scale-3 print:hidden">
    <div>
      <Link to="/">
        <img
          src="/favicon.svg"
          alt="Steampipe Logo"
          style={{ height: "28px", paddingLeft: "3px" }}
        />
      </Link>
    </div>
    <div className="w-full space-x-4 md:space-x-8 flex items-center justify-end">
      <ExternalLink
        className="text-lg text-foreground-lighter hover:text-foreground"
        withReferrer={true}
        url="https://steampipe.io/docs"
      >
        Docs
      </ExternalLink>
      <ExternalLink
        className="text-lg text-foreground-lighter hover:text-foreground"
        withReferrer={true}
        url="https://hub.steampipe.io/plugins"
      >
        Plugins
      </ExternalLink>
      <ExternalLink
        className="text-lg text-foreground-lighter hover:text-foreground"
        withReferrer={true}
        url="https://hub.steampipe.io/mods"
      >
        Mods
      </ExternalLink>
      <ThemeToggle />
    </div>
  </div>
);

export default DashboardHeader;
