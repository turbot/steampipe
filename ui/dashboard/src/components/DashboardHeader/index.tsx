import DashboardSearch from "../DashboardSearch";
import DashboardTagGroupSelect from "../DashboardTagGroupSelect";
import ExternalLink from "../ExternalLink";
import ThemeToggle from "../ThemeToggle";
import { Link } from "react-router-dom";

const DashboardHeader = () => (
  <div className="flex w-screen p-3 items-center justify-between space-x-4 bg-background-panel border-b border-table-divide print:hidden">
    <Link to="/">
      <img
          src="/favicon.svg"
          alt="Steampipe Logo"
          style={{ height: "28px", paddingLeft: "3px" }}
      />
    </Link>
    <div className="w-96">
      <DashboardSearch />
    </div>
    <DashboardTagGroupSelect />
    <div className="w-full space-x-4 md:space-x-8 flex items-center justify-end">
      <ExternalLink
        className="text-base text-foreground-lighter hover:text-foreground"
        to="https://steampipe.io/docs"
        withReferrer={true}
      >
        <>Docs</>
      </ExternalLink>
      <ExternalLink
        className="text-base text-foreground-lighter hover:text-foreground"
        to="https://hub.steampipe.io/plugins"
        withReferrer={true}
      >
        <>Plugins</>
      </ExternalLink>
      <ExternalLink
        className="text-base text-foreground-lighter hover:text-foreground"
        to="https://hub.steampipe.io/mods"
        withReferrer={true}
      >
        <>Mods</>
      </ExternalLink>
      <ThemeToggle />
    </div>
  </div>
);

export default DashboardHeader;
